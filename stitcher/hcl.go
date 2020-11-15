package stitcher

import (
	"fmt"
	"io/ioutil"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/hashicorp/hcl/v2/hclsyntax"
)

// ParseHostHCL will parse file content into valid Host.
func ParseHostHCL(src []byte, filename string) (c *Host, err error) {
	var diags hcl.Diagnostics

	file, diags := hclsyntax.ParseConfig(src, filename, hcl.Pos{Line: 1, Column: 1})
	if diags.HasErrors() {
		return nil, fmt.Errorf("Host parse: %w", diags)
	}

	c = &Host{}

	diags = gohcl.DecodeBody(file.Body, nil, c)

	if diags.HasErrors() {
		return nil, fmt.Errorf("Host parse: %w", diags)
	}

	return c, nil
}

// ReadHostHCL will load and parse a file containing hcl that defines a host
func ReadHostHCL(filename string) (c *Host, err error) {
	content, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	host, err := ParseHostHCL(content, filename)
	if err != nil {
		return nil, err
	}

	return host, nil
}
