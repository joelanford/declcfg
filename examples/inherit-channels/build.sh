#!/bin/bash

# Cleanup existing files
rm -rf tmp
rm -rf index
rm -f index.Dockerfile

# Create the raw temporary index
mkdir tmp
opm init dc-inherit-channels-operator --default-channel=stable -o yaml > tmp/tmp.yaml

yq eval-all -i '[.] | . += [{"schema":"olm.channel", "package":"dc-inherit-channels-operator", "name":"alpha"}]  | .[] | splitDoc' tmp/tmp.yaml
yq eval-all -i '[.] | . += [{"schema":"olm.channel", "package":"dc-inherit-channels-operator", "name":"beta"}]   | .[] | splitDoc' tmp/tmp.yaml
yq eval-all -i '[.] | . += [{"schema":"olm.channel", "package":"dc-inherit-channels-operator", "name":"stable"}] | .[] | splitDoc' tmp/tmp.yaml

opm render quay.io/joelanford/dc-inherit-channels-operator-bundle:v0.1.0 -o yaml >> tmp/tmp.yaml
opm render quay.io/joelanford/dc-inherit-channels-operator-bundle:v0.1.1 -o yaml >> tmp/tmp.yaml
opm render quay.io/joelanford/dc-inherit-channels-operator-bundle:v0.1.2 -o yaml >> tmp/tmp.yaml

yq eval-all -i 'select(.schema=="olm.channel" and .name=="alpha").entries  += [{"name":"dc-inherit-channels-operator.v0.1.0"}]' tmp/tmp.yaml
yq eval-all -i 'select(.schema=="olm.channel" and .name=="beta").entries   += [{"name":"dc-inherit-channels-operator.v0.1.1", "replaces": "dc-inherit-channels-operator.v0.1.0"}]' tmp/tmp.yaml
yq eval-all -i 'select(.schema=="olm.channel" and .name=="stable").entries += [{"name":"dc-inherit-channels-operator.v0.1.2", "replaces": "dc-inherit-channels-operator.v0.1.1"}]' tmp/tmp.yaml


# Build the final index with channel inheritance
mkdir index
declcfg inherit-channels tmp dc-inherit-channels-operator -o yaml > index/index.yaml
opm alpha generate dockerfile index

# Delete the tmp index
rm -rf tmp
