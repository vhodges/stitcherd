package stitcher

// RenderableContent is anything that returns a string to be merged into a document
type RenderableContent interface {
	Render(fetchcontext map[string]interface{}) (string, error)
}
