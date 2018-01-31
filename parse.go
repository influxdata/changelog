package changelog

import (
	"bytes"
	"fmt"
	"io/ioutil"

	"github.com/jsternberg/markdownfmt/markdown"
	"gopkg.in/russross/blackfriday.v2"
)

type Changelog struct {
	doc *blackfriday.Node
}

func (c *Changelog) AddEntry(e *Entry) error {
	fmt.Println(e)
	return nil
}

func (c *Changelog) WriteFile(fpath string) error {
	var buf bytes.Buffer
	renderer := markdown.NewRenderer(nil)
	renderer.RenderHeader(&buf, c.doc)
	c.doc.Walk(func(node *blackfriday.Node, entering bool) blackfriday.WalkStatus {
		return renderer.RenderNode(&buf, node, entering)
	})
	renderer.RenderFooter(&buf, c.doc)
	return ioutil.WriteFile(fpath, buf.Bytes(), 0666)
}

func ParseFile(fpath string) (*Changelog, error) {
	in, err := ioutil.ReadFile(fpath)
	if err != nil {
		return nil, err
	}

	md := blackfriday.New()
	return &Changelog{
		doc: md.Parse(in),
	}, nil
}
