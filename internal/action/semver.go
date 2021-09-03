package action

import (
	"bytes"
	"fmt"
	"html/template"
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
	SkipPatch    bool
	TemplateStrings []string
	SemverRange semver.Range
}

func (s Semver) Run() (*declcfg.DeclarativeConfig, error) {
	var out declcfg.DeclarativeConfig
	if err := copier.Copy(&out, &s.Configs); err != nil {
		return nil, fmt.Errorf("failed to copy input DC: %v", err)
	}

	// build a map of bundles in the package, mapping their name to their semver version
	versions, err := getBundleVersionsInPackage(s.PackageName, s.SemverRange, out.Bundles)
	if err != nil {
		return nil, fmt.Errorf("get bundle versions in package %q: %v", s.PackageName, err)
	}

	// generate semver-based channels, containing (still disconnected) entries
	channels, err := s.generateChannels(versions)
	if err != nil {
		return nil, fmt.Errorf("generate channels: %v", err)
	}

	// delete any channels from out that have the same name as the generated
	// channel names. we're rebuilding these channels.
	tmp := out.Channels[:0]
	for _, ch := range out.Channels {
		if _, ok := channels[ch.Name]; !ok {
			tmp = append(tmp, ch)
		}
	}
	out.Channels = tmp

	// hook up the edges for each channel, and then add the channel to out.
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
		out.Channels = append(out.Channels, *ch)
	}

	return &out, nil
}

func getBundleVersionsInPackage(pkgName string, semverRange semver.Range, allBundles []declcfg.Bundle) (map[string]semver.Version, error) {
	versions := map[string]semver.Version{}
	for _, b := range allBundles {
		if b.Package != pkgName {
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

		if semverRange(v) {
			versions[b.Name] = v
		}
	}
	return versions, nil
}

func (s Semver) generateChannels(bundles map[string]semver.Version) (map[string]*declcfg.Channel, error) {
	var tmpls []*template.Template
	for _, tStr := range s.TemplateStrings {
		t, err := template.New("").Parse(tStr)
		if err != nil {
			return nil, err
		}
		tmpls = append(tmpls, t)
	}

	channels := map[string]*declcfg.Channel{}
	for name, version := range bundles {
		for _, t := range tmpls {
			chName := mustExecuteString(t, version)
			addChannelEntry(channels, s.PackageName, chName, name)
		}
	}
	return channels, nil
}

func newChannel(pkgName, chName string) *declcfg.Channel {
	return &declcfg.Channel{
		Schema:  "olm.channel",
		Name:    chName,
		Package: pkgName,
	}
}

func addChannelEntry(channels map[string]*declcfg.Channel, pkgName, chName, entryName string) {
	ch, ok := channels[chName]
	if !ok {
		ch = newChannel(pkgName, chName)
		channels[chName] = ch
	}
	ch.Entries = append(ch.Entries, declcfg.ChannelEntry{Name: entryName})
}

func mustExecuteString(t *template.Template, data interface{}) string {
	buf := &bytes.Buffer{}
	if err := t.Execute(buf, data); err != nil {
		panic(err)
	}
	return buf.String()
}

func getMinorVersion(v semver.Version) semver.Version {
	return semver.Version{
		Major: v.Major,
		Minor: v.Minor,
	}
}