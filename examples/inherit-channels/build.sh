#!/bin/bash

# Cleanup existing files
rm -rf tmp
rm -rf index
rm -f index.Dockerfile

# Create the raw temporary index
mkdir tmp
opm init dc-inherit-channels-operator --default-channel=stable -o yaml > tmp/tmp.yaml
for b in v*; do
	opm render quay.io/joelanford/dc-inherit-channels-operator-bundle:$b -o yaml >> tmp/tmp.yaml
done

# Build the final index with channel inheritance
mkdir index
declcfg inherit-channels tmp dc-inherit-channels-operator -o yaml > index/index.yaml
opm alpha generate dockerfile index

# Delete the tmp index
rm -rf tmp
