#!/bin/bash

# Cleanup existing files
rm -rf tmp
rm -rf index index-skippatch 
rm -f index.Dockerfile index-skippatch.Dockerfile

# Create the raw temporary index
mkdir tmp 
opm init semver-operator --default-channel=stable -o yaml > tmp/tmp.yaml

yq eval-all -i '[.] | . += [{"schema":"olm.channel", "package":"semver-operator", "name":"alpha"}]  | .[] | splitDoc' tmp/tmp.yaml
yq eval-all -i '[.] | . += [{"schema":"olm.channel", "package":"semver-operator", "name":"beta"}]   | .[] | splitDoc' tmp/tmp.yaml
yq eval-all -i '[.] | . += [{"schema":"olm.channel", "package":"semver-operator", "name":"stable"}] | .[] | splitDoc' tmp/tmp.yaml

opm render quay.io/joelanford/semver-operator-bundle:v0.1.0 -o yaml >> tmp/tmp.yaml
opm render quay.io/joelanford/semver-operator-bundle:v0.1.1 -o yaml >> tmp/tmp.yaml
opm render quay.io/joelanford/semver-operator-bundle:v0.1.2 -o yaml >> tmp/tmp.yaml
opm render quay.io/joelanford/semver-operator-bundle:v0.1.3 -o yaml >> tmp/tmp.yaml

opm render quay.io/joelanford/semver-operator-bundle:v0.2.0 -o yaml >> tmp/tmp.yaml
opm render quay.io/joelanford/semver-operator-bundle:v0.2.1 -o yaml >> tmp/tmp.yaml
opm render quay.io/joelanford/semver-operator-bundle:v0.2.2 -o yaml >> tmp/tmp.yaml
opm render quay.io/joelanford/semver-operator-bundle:v0.2.3 -o yaml >> tmp/tmp.yaml

opm render quay.io/joelanford/semver-operator-bundle:v0.3.0 -o yaml >> tmp/tmp.yaml
opm render quay.io/joelanford/semver-operator-bundle:v0.3.1 -o yaml >> tmp/tmp.yaml
opm render quay.io/joelanford/semver-operator-bundle:v0.3.2 -o yaml >> tmp/tmp.yaml
opm render quay.io/joelanford/semver-operator-bundle:v0.3.3 -o yaml >> tmp/tmp.yaml

yq eval-all -i 'select(.schema=="olm.channel" and .name=="alpha").entries  += [
	{"name":"semver-operator.v0.1.0"},
	{"name":"semver-operator.v0.2.0"},
	{"name":"semver-operator.v0.1.1"},
	{"name":"semver-operator.v0.2.1"},
	{"name":"semver-operator.v0.1.2"},
	{"name":"semver-operator.v0.2.2"},
	{"name":"semver-operator.v0.1.3"},
	{"name":"semver-operator.v0.2.3"}
]' tmp/tmp.yaml
yq eval-all -i 'select(.schema=="olm.channel" and .name=="beta").entries   += [
	{"name":"semver-operator.v0.2.0"},
	{"name":"semver-operator.v0.3.0"},
	{"name":"semver-operator.v0.2.1"},
	{"name":"semver-operator.v0.3.1"},
	{"name":"semver-operator.v0.2.2"},
	{"name":"semver-operator.v0.3.2"},
	{"name":"semver-operator.v0.2.3"},
	{"name":"semver-operator.v0.3.3"}
]' tmp/tmp.yaml
yq eval-all -i 'select(.schema=="olm.channel" and .name=="stable").entries += [
	{"name":"semver-operator.v0.3.0"},
	{"name":"semver-operator.v0.3.1"},
	{"name":"semver-operator.v0.3.2"},
	{"name":"semver-operator.v0.3.3"}
]' tmp/tmp.yaml

# Build final index using semver ordering
mkdir index
declcfg semver tmp semver-operator -o yaml > index/index.yaml
opm alpha generate dockerfile index

# Build final index using semver-skippatch ordering
mkdir index-skippatch
declcfg semver tmp semver-operator --skip-patch -o yaml > index-skippatch/index.yaml
opm alpha generate dockerfile index-skippatch

# Delete the tmp index
rm -rf tmp
