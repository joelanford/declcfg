package action

import (
	"fmt"

	"github.com/jinzhu/copier"
	"github.com/operator-framework/operator-registry/alpha/declcfg"
	"github.com/operator-framework/operator-registry/alpha/property"
	"k8s.io/apimachinery/pkg/util/sets"
)

type InheritChannels struct {
	Configs     declcfg.DeclarativeConfig
	PackageName string
}

func (i InheritChannels) Run() (*declcfg.DeclarativeConfig, error) {
	return inheritChannels(i.Configs, i.PackageName)
}

func inheritChannels(in declcfg.DeclarativeConfig, packageName string) (*declcfg.DeclarativeConfig, error) {
	var out declcfg.DeclarativeConfig
	if err := copier.Copy(&out, &in); err != nil {
		return nil, fmt.Errorf("failed to copy input DC: %v", err)
	}

	bundles := map[string]bundle{}
	for i := range out.Bundles {
		if out.Bundles[i].Package != packageName {
			continue
		}

		b, err := newBundle(&out.Bundles[i])
		if err != nil {
			return nil, err
		}
		bundles[out.Bundles[i].Name] = *b
	}

	heads, err := getHeads(bundles)
	if err != nil {
		return nil, fmt.Errorf("get channel heads from bundles: %v", err)
	}

	for _, head := range heads {
		addChannelsToDescendents(bundles, head)
	}

	return &out, nil
}

func getHeads(bundles map[string]bundle) (map[string]bundle, error) {
	inChannel := map[string]sets.String{}
	replacedInChannel := map[string]sets.String{}
	skipped := sets.NewString()
	for _, b := range bundles {
		replaces := map[string]struct{}{}
		for _, ch := range b.Channels {
			replaces[ch.Replaces] = struct{}{}

			in, ok := inChannel[ch.Name]
			if !ok {
				in = sets.NewString()
			}
			in.Insert(b.Name)
			inChannel[ch.Name] = in

			rep, ok := replacedInChannel[ch.Name]
			if !ok {
				rep = sets.NewString()
			}
			rep.Insert(ch.Replaces)
			replacedInChannel[ch.Name] = rep
		}
		for _, skip := range b.Skips {
			skipped.Insert(string(skip))
		}
		if len(replaces) > 1 {
			return nil, fmt.Errorf("bundle %q has multiple replaces: channel-specific replaces not supported", b.Name)
		}
	}

	heads := map[string]bundle{}
	for name, in := range inChannel {
		replaced := replacedInChannel[name]
		chHeads := in.Difference(replaced).Difference(skipped)
		for _, h := range chHeads.List() {
			heads[h] = bundles[h]
		}
	}
	return heads, nil
}

func addChannelsToDescendents(bundleMap map[string]bundle, cur bundle) {
	for _, ch := range cur.Channels {
		next, ok := bundleMap[ch.Replaces]
		if !ok {
			continue
		}
		if len(next.Channels) == 0 {
			continue
		}
		addCh := property.Channel{Name: ch.Name, Replaces: next.Channels[0].Replaces}
		found := false
		for _, nch := range next.Channels {
			if nch == addCh {
				found = true
				break
			}
		}
		if !found {
			next.Bundle.Properties = append(next.Bundle.Properties, property.MustBuildChannel(ch.Name, next.Channels[0].Replaces))
			next.Channels = append(next.Channels, addCh)
			bundleMap[next.Name] = next
		}
		addChannelsToDescendents(bundleMap, next)
	}
}
