package stitcher

// Merge represents content and a selector string to place it at

// Replacement replaces content at At with Content
type Replacement struct {
	Content Content `hcl:"content,block"`
	At      string  `hcl:",label"`
}
