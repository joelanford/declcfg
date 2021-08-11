package action

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/jinzhu/copier"
	"github.com/operator-framework/operator-registry/alpha/declcfg"
	"github.com/operator-framework/operator-registry/alpha/property"
	"github.com/operator-framework/operator-registry/pkg/image"
	"github.com/operator-framework/operator-registry/pkg/image/containerdregistry"
	"github.com/operator-framework/operator-registry/pkg/registry"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/client-go/util/retry"
	"os"
	"regexp"
)

type InlineBundles struct {
	Configs      declcfg.DeclarativeConfig
	PackageName string

	BundleImages []string
	PruneFromNonChannelHeads bool
	Registry image.Registry

	Logger *logrus.Logger
	RegistryLogger *logrus.Logger
}

var nonRetryableRegex = regexp.MustCompile(`(error resolving name)`)

func noopLogger() *logrus.Logger {
	l := logrus.New()
	l.Out = ioutil.Discard
	return l
}

func (i InlineBundles) Run(ctx context.Context) (*declcfg.DeclarativeConfig, error) {
	var out declcfg.DeclarativeConfig
	if err := copier.Copy(&out, &i.Configs); err != nil {
		return nil, fmt.Errorf("failed to copy input DC: %v", err)
	}

	if i.Logger == nil {
		i.Logger = noopLogger()
	}
	if i.RegistryLogger == nil {
		i.RegistryLogger = noopLogger()
	}

	if i.Registry == nil {
		imageRegistry, err := containerdregistry.NewRegistry(containerdregistry.WithLog(logrus.NewEntry(i.RegistryLogger)))
		if err != nil {
			return nil, fmt.Errorf("create containerd registry: %v", err)
		}
		defer func() {
			if err := imageRegistry.Destroy(); err != nil {
				logrus.Warnf("Could not destroy containerd registry: %v", err)
			}
		}()
		i.Registry = imageRegistry
	}



	cfg := declcfg.DeclarativeConfig{}
	for _, p := range out.Packages {
		if p.Name == i.PackageName {
			cfg.Packages = append(cfg.Packages, p)
		}
	}
	for _, b := range out.Bundles {
		if b.Package == i.PackageName {
			cfg.Bundles = append(cfg.Bundles, b)
		}
	}


	bundleImages := sets.NewString(i.BundleImages...)

	allBundleImages := sets.NewString()
	for _, b := range cfg.Bundles {
		allBundleImages.Insert(b.Image)
	}
	notPresentImages := bundleImages.Difference(allBundleImages)
	if notPresentImages.Len() > 0 {
		return nil, fmt.Errorf("requested images not found: %v", notPresentImages.List())
	}

	var err error
	nonChannelHeads := sets.NewString()
	if i.PruneFromNonChannelHeads {
		nonChannelHeads, err = getAllNonChannelHeads(cfg)
		if err != nil {
			return nil, fmt.Errorf("get non-channel heads: %v", err)
		}
	}

	for j, b := range cfg.Bundles {
		blog := i.Logger.WithField("image", b.Image)
		if i.PruneFromNonChannelHeads && nonChannelHeads.Has(b.Image) {
			props := b.Properties[:0]
			for _, p := range b.Properties {
				if p.Type != property.TypeBundleObject {
					props = append(props, p)
				}
			}
			if len(props) != len(cfg.Bundles[j].Properties) {
				blog.Info("pruned olm.bundle.object properties")
			}
			cfg.Bundles[j].Properties = props
			blog.Info("skipping non-channel head")
		} else if bundleImages.Len() == 0 || bundleImages.Has(b.Image) {
			imgRef := image.SimpleReference(b.Image)

			if err := retry.OnError(retry.DefaultRetry,
				func(err error) bool {
					if nonRetryableRegex.MatchString(err.Error()) {
						return false
					}
					blog.Warnf("  Error pulling image: %v. Retrying.", err)
					return true
				},
				func() error { return i.Registry.Pull(ctx, imgRef) },
			); err != nil {
				return nil, fmt.Errorf("pull image %q: %v", imgRef, err)
			}

			tmpDir, err := os.MkdirTemp("", "declcfg-inline-bundles-")
			if err != nil {
				return nil, err
			}
			if err := i.Registry.Unpack(ctx, imgRef, tmpDir); err != nil {
				return nil, err
			}
			ii, err := registry.NewImageInput(image.SimpleReference(b.Image), tmpDir)
			if err != nil {
				return nil, err
			}
			props := b.Properties[:0]
			for _, p := range b.Properties {
				if p.Type != property.TypeBundleObject {
					props = append(props, p)
				}
			}

			for _, obj := range ii.Bundle.Objects {
				objJson, err := json.Marshal(obj)
				if err != nil {
					return nil, err
				}
				props = append(props, property.MustBuildBundleObjectData(objJson))
			}
			b.Properties = props
			cfg.Bundles[j] = b
			blog.Info("inlined olm.bundle.object properties")
		}
	}

	pi := 0
	for j, p := range out.Packages {
		if p.Name == i.PackageName {
			out.Packages[j] = cfg.Packages[pi]
			pi++
		}
	}
	bi := 0
	for j, b := range out.Bundles {
		if b.Package == i.PackageName {
			out.Bundles[j] = cfg.Bundles[bi]
			bi++
		}
	}

	return &out, nil
}

func getAllNonChannelHeads(cfg declcfg.DeclarativeConfig) (sets.String, error) {
	m, err := declcfg.ConvertToModel(cfg)
	if err != nil {
		return nil, fmt.Errorf("convert index to model: %v", err)
	}

	nonChannelHeads := sets.NewString()
	for _, pkg := range m {
		for _, ch := range pkg.Channels {
			for _, b := range ch.Bundles {
				nonChannelHeads.Insert(b.Image)
			}
		}
	}
	for _, pkg := range m {
		for _, ch := range pkg.Channels {
			head, err := ch.Head()
			if err != nil {
				return nil, err
			}
			nonChannelHeads.Delete(head.Image)
		}
	}
	return nonChannelHeads, nil
}
