package main

import (
	"github.com/vhodges/stitcherd/cmd"
	"github.com/vhodges/stitcherd/stitcher"

	"fmt"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/hashicorp/hcl/v2/hclsyntax"
)

// ParseHost will parse file content into valid Host.
func ParseHost(src []byte, filename string) (c *stitcher.Host, err error) {
	var diags hcl.Diagnostics

	file, diags := hclsyntax.ParseConfig(src, filename, hcl.Pos{Line: 1, Column: 1})
	if diags.HasErrors() {
		return nil, fmt.Errorf("Host parse: %w", diags)
	}

	c = &stitcher.Host{}

	diags = gohcl.DecodeBody(file.Body, nil, c)

	if diags.HasErrors() {
		return nil, fmt.Errorf("Host parse: %w", diags)
	}

	return c, nil
}

func main() {
	cmd.Execute()
}
