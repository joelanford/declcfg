#!/bin/bash

# Cleanup existing files
rm -rf tmp
rm -rf index index-skippatch 
rm -f index.Dockerfile index-skippatch.Dockerfile

# Create the raw temporary index
mkdir tmp 
opm init semver-operator --default-channel=default -o yaml > tmp/index.yaml

for v in v*; do
	opm render quay.io/joelanford/semver-operator-bundle:$v -o yaml >> tmp/index.yaml
done

# Build final index using semver ordering
cp -r tmp index
echo "$(declcfg semver index semver-operator -t "default" -o yaml)" > index/index.yaml
echo "$(declcfg semver index semver-operator -r ">=0.3.0" -t "stable-v{{.Major}}.{{.Minor}}" -o yaml)" > index/index.yaml
echo "$(declcfg semver index semver-operator -r ">0.1.x <0.3.0" -t "semver-operator-v{{.Major}}.{{.Minor}}" -o yaml)" > index/index.yaml
opm alpha generate dockerfile index

# Build final index using semver-skippatch ordering
cp -r tmp index-skippatch
echo "$(declcfg semver index-skippatch semver-operator -t "default" --skip-patch -o yaml)" > index-skippatch/index.yaml
echo "$(declcfg semver index-skippatch semver-operator -r ">=0.3.0" -t "stable-v{{.Major}}.{{.Minor}}" --skip-patch -o yaml)" > index-skippatch/index.yaml
echo "$(declcfg semver index-skippatch semver-operator -r ">0.1.x <0.3.0" -t "semver-operator-v{{.Major}}.{{.Minor}}" --skip-patch -o yaml)" > index-skippatch/index.yaml
opm alpha generate dockerfile index-skippatch

# Delete the tmp index
rm -rf tmp