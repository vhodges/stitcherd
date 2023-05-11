
package stitcher

import (
	"log"
	"github.com/PuerkitoBio/goquery"
)

type DocumentTransform struct {
	Type string
	ParentSelector string
	ChildSelector string

	Classname string 
}

func (transform *DocumentTransform) Transform(parent_doc *goquery.Document, child_doc *goquery.Document) {

	switch transform.Type {
	case "replace":

		if parent_doc == nil || transform.ParentSelector == "" || child_doc == nil {
			return
		}

		replaceAt := parent_doc.Find(transform.ParentSelector)

		if transform.ChildSelector != "" {
			replaceWith := child_doc.Find(transform.ChildSelector)
			html, err := replaceWith.Html()
			if err != nil {
				log.Printf("err: '%v'\n", err)
				return
			}
			replaceAt.ReplaceWithHtml(html)
		} else {
			html, err := child_doc.Html()
			if err != nil {
				log.Printf("err: '%v'\n", err)
				return
			}
			replaceAt.ReplaceWithHtml(html)
		}
	case "set_class":
		if parent_doc == nil || transform.ParentSelector == "" || transform.Classname == "" {
			return
		}
		// TODO
	default:
	}
}
