package action

import (
	"fmt"
	"sort"

	"github.com/blang/semver/v4"
	"github.com/jinzhu/copier"
	"github.com/operator-framework/operator-registry/alpha/declcfg"
	"github.com/operator-framework/operator-registry/alpha/property"
	"k8s.io/apimachinery/pkg/util/sets"
)

type Semver struct {
	Configs      declcfg.DeclarativeConfig
	PackageName  string
	ChannelNames []string
	SkipPatch    bool
}

func (s Semver) Run() (*declcfg.DeclarativeConfig, error) {
	var out declcfg.DeclarativeConfig
	if err := copier.Copy(&out, &s.Configs); err != nil {
		return nil, fmt.Errorf("failed to copy input DC: %v", err)
	}

	channels := []*declcfg.Channel{}
	channelNameSet := sets.NewString(s.ChannelNames...)
	foundChannels := sets.NewString()
	entries := sets.NewString()
	for _, ch := range out.Channels {
		ch := ch
		if ch.Package != s.PackageName {
			continue
		}
		if channelNameSet.Len() == 0 ||  channelNameSet.Has(ch.Name) {
			channels = append(channels, &ch)
			foundChannels = foundChannels.Insert(ch.Name)
			for _, e := range ch.Entries {
				entries = entries.Insert(e.Name)
			}
		}
	}

	missingChannels := channelNameSet.Difference(foundChannels)
	if missingChannels.Len() > 0  {
		return nil, fmt.Errorf("could not find specified channels: %v", missingChannels.List())
	}

	versions := map[string]semver.Version{}
	for _, b := range out.Bundles {
		if b.Package != s.PackageName || !entries.Has(b.Name) {
			continue
		}
		props, err := property.Parse(b.Properties)
		if err != nil {
			return nil, fmt.Errorf("parse properties for bundle %q: %v", b.Name, err)
		}
		if len(props.Packages) != 1 {
			return nil, fmt.Errorf("bundle %q has multiple %q properties, expected exactly 1", b.Name, property.TypePackage)
		}
		v, err := semver.Parse(props.Packages[0].Version)
		if err != nil {
			return nil, fmt.Errorf("bundle %q has invalid version %q: %v", b.Name, props.Packages[0].Version, err)
		}
		versions[b.Name] = v
	}

	for _, ch := range channels {
		ch := ch
		sort.Slice(ch.Entries, func(i,j int) bool {
			return versions[ch.Entries[i].Name].LT(versions[ch.Entries[j].Name])
		})

		if !s.SkipPatch {
			for i := 1; i < len(ch.Entries); i++ {
				ch.Entries[i] = declcfg.ChannelEntry{
					Name:      ch.Entries[i].Name,
					Replaces:  ch.Entries[i-1].Name,
				}
			}
		} else {
			curIndex := len(ch.Entries)-1
			curMinor := getMinorVersion(versions[ch.Entries[curIndex].Name])
			curSkips := sets.NewString()
			for i := len(ch.Entries)-2; i >=0; i-- {
				thisName := ch.Entries[i].Name
				thisMinor := getMinorVersion(versions[thisName])
				if thisMinor.EQ(curMinor) {
					ch.Entries[i] = declcfg.ChannelEntry{Name: thisName}
					curSkips = curSkips.Insert(thisName)
				} else {
					ch.Entries[curIndex] = declcfg.ChannelEntry{
						Name:      ch.Entries[curIndex].Name,
						Replaces:  thisName,
						Skips:     curSkips.List(),
					}
					curSkips = sets.NewString()
					curIndex = i
					curMinor = thisMinor
				}
			}
			ch.Entries[curIndex] = declcfg.ChannelEntry{
				Name:      ch.Entries[curIndex].Name,
				Skips:     curSkips.List(),
			}
		}
	}

	return &out, nil
}

func getMinorVersion(v semver.Version) semver.Version {
	return semver.Version{
		Major: v.Major,
		Minor: v.Minor,
	}
}