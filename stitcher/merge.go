package stitcher

// Merge represents content and a selector string to place it at

// Merge replaces content at At with Content
type Merge struct {
	Content EndPoint `hcl:"with,block"`
	At      string   `hcl:",label"`
}
