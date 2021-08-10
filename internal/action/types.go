package action

import (
	"fmt"

	"github.com/operator-framework/operator-registry/alpha/declcfg"
	"github.com/operator-framework/operator-registry/alpha/property"
)

type bundle struct {
	*declcfg.Bundle
	property.Properties
}

func newBundle(in *declcfg.Bundle) (*bundle, error) {
	props, err := property.Parse(in.Properties)
	if err != nil {
		return nil, fmt.Errorf("parse properties for bundle %q", in.Name)
	}
	return &bundle{
		Bundle:     in,
		Properties: *props,
	}, nil
}
