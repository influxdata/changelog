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
	reVersionHeader = regexp.MustCompile(`^v(\d+(\.\d+)*) \[(.*)\]$`)
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

func (c *Changelog) AddEntry(e *Entry) error {
	// Find the first version heading.
	var versionHeader *blackfriday.Node
	c.doc.Walk(func(node *blackfriday.Node, entering bool) blackfriday.WalkStatus {
		switch node.Type {
		case blackfriday.Heading:
			if node.HeadingData.Level == 2 {
				m := reVersionHeader.FindStringSubmatch(getHeadingText(node))
				if m == nil {
					return blackfriday.GoToNext
				} else if m[3] == "unreleased" {
					versionHeader = node
				} else {
					parts := strings.Split(m[1], ".")
					indices := make([]int, len(parts))
					for i, s := range parts {
						indices[i], _ = strconv.Atoi(s)
					}

					if len(indices) >= 2 {
						indices[1]++
					}
					for len(indices) < 3 {
						indices = append(indices, 0)
					}

					if len(parts) < len(indices) {
						parts = make([]string, len(indices))
					}
					for i, n := range indices {
						parts[i] = strconv.Itoa(n)
					}
					version := strings.Join(parts, ".")

					versionHeader = blackfriday.NewNode(blackfriday.Heading)
					versionHeader.HeadingData.Level = 2
					headingText := blackfriday.NewNode(blackfriday.Text)
					headingText.Literal = []byte(fmt.Sprintf("v%s [unreleased]", version))
					versionHeader.AppendChild(headingText)
					node.InsertBefore(versionHeader)
				}
				return blackfriday.Terminate
			}
		}
		return blackfriday.GoToNext
	})

	// Include the first version heading.
	if versionHeader == nil {
		versionHeader = blackfriday.NewNode(blackfriday.Heading)
		versionHeader.HeadingData.Level = 2
		headingText := blackfriday.NewNode(blackfriday.Text)
		headingText.Literal = []byte("v1.0.0 [unreleased]")
		versionHeader.AppendChild(headingText)
		c.doc.AppendChild(versionHeader)
	}

	// Within this version heading, attempt to find a heading that matches the section we want to add.
	var section *blackfriday.Node
	headingData, ok := headings[e.Type]
	if !ok {
		return nil
	}
	for n := versionHeader.Next; n != nil; n = n.Next {
		if n.Type != blackfriday.Heading {
			continue
		}
		text := getHeadingText(n)
		if n.HeadingData.Level <= 2 || (headingData.IsBefore != "" && text == headingData.IsBefore) {
			// We need to insert this heading here.
			section = blackfriday.NewNode(blackfriday.Heading)
			section.HeadingData.Level = 3
			headingText := blackfriday.NewNode(blackfriday.Text)
			headingText.Literal = []byte(headingData.Name)
			section.AppendChild(headingText)
			n.InsertBefore(section)
			break
		} else if n.HeadingData.Level == 3 && text == headingData.Name {
			section = n
			break
		}
	}

	if section == nil {
		section = blackfriday.NewNode(blackfriday.Heading)
		section.HeadingData.Level = 3
		headingText := blackfriday.NewNode(blackfriday.Text)
		headingText.Literal = []byte(headingData.Name)
		section.AppendChild(headingText)
		c.doc.AppendChild(section)
	}

	// Insert a list as the first element after the heading if it does not exist.
	list := section.Next
	if list == nil || list.Type != blackfriday.List {
		list = blackfriday.NewNode(blackfriday.List)
		list.ListData = blackfriday.ListData{
			Tight:      true,
			BulletChar: '-',
		}

		if section.Next != nil {
			section.Next.InsertBefore(list)
		} else {
			section.Parent.AppendChild(list)
		}
	}

	// Create the list item and insert it at the end.
	linkText := blackfriday.NewNode(blackfriday.Text)
	linkText.Literal = []byte(fmt.Sprintf("#%d", e.Number))
	link := blackfriday.NewNode(blackfriday.Link)
	link.LinkData = blackfriday.LinkData{
		Destination: []byte(e.URL.String()),
	}
	link.AppendChild(linkText)
	comment := blackfriday.NewNode(blackfriday.Text)
	comment.Literal = []byte(fmt.Sprintf(": %s.", e.Message))

	paragraph := blackfriday.NewNode(blackfriday.Paragraph)
	paragraph.AppendChild(link)
	paragraph.AppendChild(comment)

	item := blackfriday.NewNode(blackfriday.Item)
	item.AppendChild(paragraph)
	if list.FirstChild == nil {
		item.ListFlags = blackfriday.ListItemBeginningOfList
	}
	list.AppendChild(item)
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

	md := blackfriday.New(blackfriday.WithExtensions(blackfriday.CommonExtensions))
	return &Changelog{
		doc: md.Parse(in),
	}, nil
}

func getHeadingText(node *blackfriday.Node) string {
	var buf bytes.Buffer
	for n := node.FirstChild; n != nil; n = n.Next {
		if n.Type == blackfriday.Text {
			buf.Write(n.Literal)
		}
	}
	return buf.String()
}
