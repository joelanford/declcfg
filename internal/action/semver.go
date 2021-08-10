package action

import (
	"fmt"
	"sort"

	"github.com/blang/semver/v4"
	"github.com/jinzhu/copier"
	"github.com/operator-framework/operator-registry/alpha/declcfg"
	"github.com/operator-framework/operator-registry/alpha/property"
	"k8s.io/apimachinery/pkg/util/json"
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

	// declare a map of channels, where values are lists of bundle pointers.
	// pointers for bundles are important so that updates made for specific
	// channel processing are reflected across all channel map values.
	channels := map[string][]*bundle{}

	// Build a map of the versions for quick lookups by bundle name.
	versions := map[string]semver.Version{}
	for i := range out.Bundles {
		// ignore bundles for other packages
		if out.Bundles[i].Package != s.PackageName {
			continue
		}

		// create a new bundle instance (containing parsed properties)
		b, err := newBundle(&out.Bundles[i])
		if err != nil {
			return nil, err
		}

		// populate versions map.
		if len(b.Packages) != 1 {
			return nil, fmt.Errorf("bundle %q must have exactly one package property, found %d", b.Name, len(b.Packages))
		}
		v, err := semver.Parse(b.Packages[0].Version)
		if err != nil {
			return nil, fmt.Errorf("bundle %q has invalid version %q", b.Name, b.Packages[0].Version)
		}
		versions[b.Name] = v

		// if the requested channel name list is non-empty, add this bundle to
		// the channel map for each of the requested channels that it is a
		// member of.
		for _, ch := range s.ChannelNames {
			if inChannel(*b, ch) {
				channels[ch] = append(channels[ch], b)
			}
		}

		// if the requested channel name list is empty, add this bundle to the
		// channel map for all channels it is a member of.
		if len(s.ChannelNames) == 0 {
			for _, ch := range b.Channels {
				channels[ch.Name] = append(channels[ch.Name], b)
			}
		}
	}

	// Handle each channel separately
	for chName, bundles := range channels {
		// If the channel has only one bundle, there's nothing to do.
		if len(bundles) <= 1 {
			continue
		}

		// Sort the bundles by version, ascending.
		sort.Slice(bundles, func(i, j int) bool {
			return versions[bundles[i].Name].LT(versions[bundles[j].Name])
		})

		if !s.SkipPatch {
			// Iterate the sorted bundles, and change the channel property for the
			// channel we're working on such that each bundle replaces the one
			// before it in the list (and such that the first bundle in the list
			// replaces nothing)
			for i := 0; i < len(bundles); i++ {
				for j, p := range bundles[i].Bundle.Properties {
					if p.Type != property.TypeChannel {
						continue
					}
					var ch property.Channel
					// Not necessary to check this error because we've already
					// parsed ALL of the properties successfully.
					_ = json.Unmarshal(p.Value, &ch)
					if chName == ch.Name {
						replaces := ""
						if i > 0 {
							replaces = bundles[i-1].Name
						}
						bundles[i].Bundle.Properties[j] = property.MustBuildChannel(ch.Name, replaces)
					}
				}
			}
		} else {
			// if skipPatch is enabled, ensure that a skip properties exists to
			// enable upgrades to skip over intermediate patch releases.
			curReplace := bundles[0].Name
			nextReplace := ""
			curSkips := sets.NewString()
			curY := versions[bundles[0].Name].Minor
			for i := 0; i < len(bundles); i++ {
				thisY := versions[bundles[i].Name].Minor
				if curY == thisY {
					// bundles[i] should replace curReplace in this channel (or
					// nothing if i == 0)
					for j, p := range bundles[i].Bundle.Properties {
						if p.Type != property.TypeChannel {
							continue
						}
						var ch property.Channel
						// Not necessary to check this error because we've already
						// parsed ALL of the properties successfully.
						_ = json.Unmarshal(p.Value, &ch)
						if chName == ch.Name {
							replaces := ""
							if i > 0 {
								replaces = curReplace
							}
							bundles[i].Bundle.Properties[j] = property.MustBuildChannel(ch.Name, replaces)
						}
					}

					// bundles[i] should skip everything back to but not including curReplace
					// TODO(joelanford): Should we clean up other skips values that were already
					//     predefined?
					have := sets.NewString()
					for _, s := range bundles[i].Skips {
						have = have.Insert(string(s))
					}
					for _, s := range curSkips.Difference(have).List() {
						bundles[i].Bundle.Properties = append(bundles[i].Bundle.Properties, property.MustBuildSkips(s))
						bundles[i].Skips = append(bundles[i].Skips, property.Skips(s))
					}

					// bundles[i] should be added to the current set of skips so that future bundles in this Y stream
					// skip it.
					if bundles[i].Name != curReplace {
						curSkips = curSkips.Insert(bundles[i].Name)
					}

					// bump nextReplace so that when we finally jump to the next Y stream, nextReplace is set to the
					// highest semver bundle from the previous Y stream.
					nextReplace = bundles[i].Name
				} else {
					// we found a new minor version, so set curReplace to nextReplace, update curY, and reset curSkips.
					// curSkips should include this bundle so that other bundles in this Y series skip it.
					curReplace = nextReplace
					curY = thisY
					curSkips = sets.NewString(bundles[i].Name)

					// we need to replace the new curReplace in this channel.
					// NOTE: since this is the first version of a minor release, we should not add any additional skips
					for j, p := range bundles[i].Bundle.Properties {
						if p.Type != property.TypeChannel {
							continue
						}
						var ch property.Channel
						// Not necessary to check this error because we've already
						// parsed ALL of the properties successfully.
						_ = json.Unmarshal(p.Value, &ch)
						if chName == ch.Name {
							bundles[i].Bundle.Properties[j] = property.MustBuildChannel(ch.Name, curReplace)
						}
					}
				}
			}
		}
	}

	return &out, nil
}

func inChannel(b bundle, channelName string) bool {
	for _, ch := range b.Channels {
		if ch.Name == channelName {
			return true
		}
	}
	return false
}
