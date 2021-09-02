package action

import (
	"fmt"

	"github.com/jinzhu/copier"
	"github.com/operator-framework/operator-registry/alpha/declcfg"
	"k8s.io/apimachinery/pkg/util/sets"
)

type InheritChannels struct {
	Configs     declcfg.DeclarativeConfig
	PackageName string
}

func (inherit InheritChannels) Run() (*declcfg.DeclarativeConfig, error) {
	var out declcfg.DeclarativeConfig
	if err := copier.Copy(&out, &inherit.Configs); err != nil {
		return nil, fmt.Errorf("failed to copy input DC: %v", err)
	}

	// allReplaces tracks every entry each bundle replaces across all channels
	// inherit-channels is only supported when no bundle replaces two or more
	// different bundles globally, so we use this map to verify this requirement
	allReplaces := map[string]sets.String{}

	// replaces tracks the replaces value for each bundle. if the allReplaces
	// bundle check described above passes, the replaces map will tell us what
	// the globally replaced bundle is for each bundle.
	replaces := map[string]string{}

	// entryChannels maps a bundle to the set of channels it is a member of.
	entryChannels := map[string]sets.String{}
	for _, channel := range out.Channels {
		if channel.Package != inherit.PackageName {
			continue
		}
		for _, e := range channel.Entries {
			if _, ok := allReplaces[e.Name]; !ok {
				allReplaces[e.Name] = sets.NewString()
			}
			allReplaces[e.Name] = allReplaces[e.Name].Insert(e.Replaces)
			replaces[e.Name] = e.Replaces

			if _, ok := entryChannels[e.Name]; !ok {
				entryChannels[e.Name] = sets.NewString()
			}
			entryChannels[e.Name] = entryChannels[e.Name].Insert(channel.Name)
		}
	}
	for name, entryReplacesSet := range allReplaces {
		if entryReplacesSet.Len() < 1{
			return nil, fmt.Errorf("entry %q has multiple replaces (%v): channel-specific replaces not supported", name, entryReplacesSet.List())
		}
	}

	// now that we know all of the channels of all of the entries, it is
	// just a matter of finding bundles that replace other bundles that are
	// in the parent bundle's channels.
	newEntries := map[string][]declcfg.ChannelEntry{}
	for _, ch := range out.Channels {
		for _, e := range ch.Entries {
			curChannels := entryChannels[e.Name]
			replacedChannels, ok := entryChannels[e.Replaces]
			inheritChannels := sets.NewString()
			if ok {
				inheritChannels = curChannels.Difference(replacedChannels)
			}
			// if replacedChannels is missing some channels present in
			// curChannels, then we've found a tail bundle that references an
			// entry in another channel.
			//
			// step through the replaces chain from
			// here and add new entries in the current channel for every bundle
			// in the chain.
			if inheritChannels.Len() > 0 {
				cur := e.Replaces
				for cur != "" {
					replace := replaces[cur]
					newEntries[ch.Name] = append(newEntries[ch.Name], declcfg.ChannelEntry{
						Name:      cur,
						Replaces:  replace,
					})
					cur = replace
				}
			}
		}
	}

	// add all of the new entries for each channel
	for i := range out.Channels {
		out.Channels[i].Entries = append(out.Channels[i].Entries, newEntries[out.Channels[i].Name]...)
	}

	return &out, nil
}