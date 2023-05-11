package stitcher

import (
)

// Yes, this duplicates Fragment but simplifies marshalling
type FragmentedPage struct {
	Fragment Fragment
}

func (page *FragmentedPage) Render(site *Host, contextdata map[string]interface{}) string {
	var err error
	var content string = ""

	if page.Fragment.Cachable() {
		content, err = page.Fragment.FromCache(site, contextdata)
		if err != nil {
			return "<!-- FRAGMENT ERROR -->" // TODO Make this better
		}
	} else {
		content = page.Fragment.Render(site, contextdata)		
	}

	return content
}
