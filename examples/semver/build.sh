#!/bin/bash

# Cleanup existing files
rm -rf tmp
rm -rf index index-skippatch 
rm -f index.Dockerfile index-skippatch.Dockerfile

# Create the raw temporary index
mkdir tmp 
opm init semver-operator --default-channel=stable -o yaml > tmp/tmp.yaml
for b in v*; do
	opm render quay.io/joelanford/semver-operator-bundle:$b -o yaml >> tmp/tmp.yaml
done

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
