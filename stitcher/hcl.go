package stitcher

import (
	"fmt"
	"io/ioutil"
	"log"
	"unicode/utf8"

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

// From fs.  Will go away once updated to Go 1.16.x
func ValidPath(name string) bool {
	if !utf8.ValidString(name) {
		return false
	}

	if name == "." {
		// special case
		return true
	}

	// Iterate over elements in name, checking each.
	for {
		i := 0
		for i < len(name) && name[i] != '/' {
			i++
		}
		elem := name[:i]
		if elem == "" || elem == "." || elem == ".." {
			return false
		}
		if i == len(name) {
			return true // reached clean ending
		}
		name = name[i+1:]
	}
}

// ReadHostHCL will load and parse a file containing hcl that defines a host
func ReadHostHCL(filename string) (c *Host, err error) {

	if !ValidPath(filename) {
		log.Printf("Invalid filename '%s'\n", filename)
		return nil, fmt.Errorf("Invalid filename '%s'", filename)
	}
	
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
