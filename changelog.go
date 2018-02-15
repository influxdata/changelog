package changelog

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"regexp"
	"strconv"
	"strings"

	"github.com/jsternberg/markdownfmt/markdown"
	"gopkg.in/russross/blackfriday.v2"
)

var (
	reVersionHeader = regexp.MustCompile(`^v(\d+(\.\d+)*) \[(.*)]$`)
	headings        = map[EntryType]struct {
		Name string
		// IsBefore marks which heading entry this entry is before. That way
		// we can ensure that it is inserted within the correct place in the document
		// and we can stop searching if we see this section.
		// The entry should be inserted at the last possible location that is before this
		// section.
		IsBefore string
	}{
		FeatureRequest: {
			Name:     "Features",
			IsBefore: "Bugfixes",
		},
		Bugfix: {
			Name: "Bugfixes",
		},
	}
)

type Changelog struct {
	doc *blackfriday.Node
}

func New() *Changelog {
	doc := blackfriday.NewNode(blackfriday.Document)
	return &Changelog{doc: doc}
}

func (c *Changelog) AddEntry(e *Entry) {
	heading, ok := headings[e.Type]
	if !ok {
		return
	}

	section := c.findOrCreateHeading(nil, 0, func(text string) int {
		m := reVersionHeader.FindStringSubmatch(text)
		if m == nil {
			return -1
		}

		ver, err := NewVersion(m[1])
		if err != nil {
			return -1
		}
		return e.Version.Compare(ver)
	}, func() *blackfriday.Node {
		text := blackfriday.NewNode(blackfriday.Text)
		text.Literal = []byte(fmt.Sprintf("v%s [unreleased]", e.Version))
		node := blackfriday.NewNode(blackfriday.Heading)
		node.HeadingData.Level = 2
		node.AppendChild(text)
		return node
	})

	// Look for the entry within this section. Take a look at every list and try to match the pattern.
	// The section does not matter for the purposes of this function since we trust what is in the file
	// more than the metadata we have been given.
	for n := section.Next; n != nil; n = n.Next {
		if n.Type == blackfriday.Heading && n.HeadingData.Level <= section.Level {
			// We have entered a new section. We are looking for this entry only in this version.
			// The entry being present in other versions does not matter.
			break
		}

		found := false
		n.Walk(func(node *blackfriday.Node, entering bool) blackfriday.WalkStatus {
			switch node.Type {
			case blackfriday.Link:
				var buf bytes.Buffer
				for n := node.FirstChild; n != nil; n = n.Next {
					n.Walk(func(node *blackfriday.Node, entering bool) blackfriday.WalkStatus {
						buf.Write(node.Literal)
						return blackfriday.GoToNext
					})
				}

				text := buf.Bytes()
				text = bytes.TrimPrefix(text, []byte{'#'})
				n, err := strconv.Atoi(string(text))
				if err != nil {
					return blackfriday.SkipChildren
				}

				if n == e.Number {
					found = true
					return blackfriday.Terminate
				}
				return blackfriday.SkipChildren
			}
			return blackfriday.GoToNext
		})

		if found {
			// There is no need to insert this entry since it has been found.
			return
		}
	}

	// Find the section we wish to insert this value into.
	entries := c.findOrCreateHeading(section, section.HeadingData.Level, func(text string) int {
		if text == heading.Name {
			return 0
		} else if text == heading.IsBefore {
			return 1
		}
		return -1
	}, func() *blackfriday.Node {
		text := blackfriday.NewNode(blackfriday.Text)
		text.Literal = []byte(heading.Name)
		node := blackfriday.NewNode(blackfriday.Heading)
		node.HeadingData.Level = section.HeadingData.Level + 1
		node.AppendChild(text)
		return node
	})

	// If the first element after the heading is a list, insert it at the end of the list. If it isn't
	// a list, insert a list.
	var list *blackfriday.Node
	if entries.Next != nil && entries.Next.Type == blackfriday.List {
		list = entries.Next
	} else {
		list = blackfriday.NewNode(blackfriday.List)
		list.ListData = blackfriday.ListData{
			Tight:      true,
			BulletChar: '-',
		}
		if entries.Next != nil {
			entries.Next.InsertBefore(list)
		} else {
			entries.Parent.AppendChild(list)
		}
	}

	// Add an item to the end of the list. We have already checked if the list item exists and it doesn't,
	// so feel safe adding it.
	item := c.createListItem(e)
	if list.LastChild != nil {
		list.LastChild.ListFlags &= ^blackfriday.ListItemEndOfList
	}
	if list.FirstChild == nil {
		item.ListFlags |= blackfriday.ListItemBeginningOfList
	}
	item.ListFlags |= blackfriday.ListItemEndOfList
	list.AppendChild(item)
}

func (c *Changelog) createListItem(e *Entry) *blackfriday.Node {
	text := blackfriday.NewNode(blackfriday.Text)
	text.Literal = []byte(fmt.Sprintf("#%d", e.Number))
	link := blackfriday.NewNode(blackfriday.Link)
	link.LinkData = blackfriday.LinkData{
		Destination: []byte(e.URL.String()),
	}
	link.AppendChild(text)
	comment := blackfriday.NewNode(blackfriday.Text)
	message := e.Message
	if !hasPunctuation(message) {
		message += "."
	}
	comment.Literal = []byte(fmt.Sprintf(": %s", message))

	paragraph := blackfriday.NewNode(blackfriday.Paragraph)
	paragraph.AppendChild(link)
	paragraph.AppendChild(comment)

	item := blackfriday.NewNode(blackfriday.Item)
	item.AppendChild(paragraph)
	return item
}

func (c *Changelog) findOrCreateHeading(start *blackfriday.Node, level int, cmp func(text string) int, create func() *blackfriday.Node) (section *blackfriday.Node) {
	if start == nil {
		if c.doc.FirstChild == nil {
			if create != nil {
				section = create()
				c.doc.AppendChild(section)
			}
			return section
		}
		start = c.doc.FirstChild
	} else if start.Next == nil {
		if create != nil {
			section = create()
			start.Parent.AppendChild(section)
		}
		return section
	} else {
		start = start.Next
	}

	for n := start; n != nil; n = n.Next {
		n.Walk(func(node *blackfriday.Node, entering bool) blackfriday.WalkStatus {
			switch node.Type {
			case blackfriday.Heading:
				// If the heading is outside of the range we are looking for, then try to create the node and terminate.
				if node.HeadingData.Level <= level {
					if create != nil {
						section = create()
						node.InsertBefore(section)
						return blackfriday.Terminate
					}
				}

				var buf bytes.Buffer
				for n := node.FirstChild; n != nil; n = n.Next {
					n.Walk(func(node *blackfriday.Node, entering bool) blackfriday.WalkStatus {
						buf.Write(node.Literal)
						return blackfriday.GoToNext
					})
				}

				if val := cmp(buf.String()); val == 0 {
					section = node
					return blackfriday.Terminate
				} else if val > 0 {
					if create != nil {
						section = create()
						node.InsertBefore(section)
					}
					return blackfriday.Terminate
				}
				return blackfriday.SkipChildren
			}
			return blackfriday.GoToNext
		})

		if section != nil {
			return section
		}
	}

	if section == nil && create != nil {
		section = create()
		start.Parent.AppendChild(section)
	}
	return section
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

func Parse(in []byte) *Changelog {
	md := blackfriday.New(blackfriday.WithExtensions(blackfriday.CommonExtensions))
	return &Changelog{
		doc: md.Parse(in),
	}
}

func ParseFile(fpath string) (*Changelog, error) {
	in, err := ioutil.ReadFile(fpath)
	if err != nil {
		return nil, err
	}
	return Parse(in), nil
}

func hasPunctuation(msg string) bool {
	for _, s := range []string{".", "!", "?"} {
		if strings.HasSuffix(msg, s) {
			return true
		}
	}
	return false
}
