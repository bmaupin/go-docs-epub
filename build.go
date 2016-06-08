package main

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	"golang.org/x/net/html"

	"github.com/bmaupin/go-epub"
	"github.com/bmaupin/go-util/htmlutil"
)

const (
	effectiveGoCoverImg      = "covers/effective-go.png"
	effectiveGoFilename      = "Effective Go.epub"
	effectiveGoSectionTag    = "h2"
	effectiveGoSeparator     = "<h2"
	effectiveGoTitle         = "Effective Go"
	effectiveGoTitleFilename = "title.xhtml"
	effectiveGoUrl           = "https://golang.org/doc/effective_go.html"
)

type epubSection struct {
	title    string
	filename string
	nodes    []html.Node
}

func main() {
	err := buildEffectiveGo()
	if err != nil {
		log.Println("Error building Effective Go")
	}
}

func buildEffectiveGo() error {
	resp, err := http.Get(effectiveGoUrl)
	if err != nil {
		return err
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			panic(err)
		}
	}()

	doc, err := html.Parse(resp.Body)
	if err != nil {
		log.Fatalf("Parse error: %s", err)
	}

	pageNode := htmlutil.GetFirstHtmlNode(doc, "div", "id", "page")
	containerNode := htmlutil.GetFirstHtmlNode(pageNode, "div", "class", "container")

	footerNode := htmlutil.GetFirstHtmlNode(containerNode, "div", "id", "footer")
	footerNode = reformatEffectiveGoFooter(footerNode)

	sections := []epubSection{}
	sectionFilename := effectiveGoTitleFilename
	// Don't add a title so it doesn't get added to the TOC
	section := &epubSection{
		filename: effectiveGoTitleFilename,
	}
	internalLinks := make(map[string]string)

	// Iterate through each child node
	for node := containerNode.FirstChild; node != nil; node = node.NextSibling {
		// If we find the section tag
		if node.Type == html.ElementNode && node.Data == effectiveGoSectionTag {
			// Add the previous section to the slice of sections
			sections = append(sections, *section)

			sectionTitle := node.FirstChild.Data
			sectionFilename = titleToFilename(sectionTitle)

			// Start a new section
			section = &epubSection{
				filename: sectionFilename,
				title:    sectionTitle,
			}
		}

		if node == footerNode {
			// Add the footer node to the title page since it contains license information
			sections[0].nodes = append(sections[0].nodes, *node)
		} else {
			// Append the current node to the current section
			section.nodes = append(section.nodes, *node)
		}

		for _, n := range htmlutil.GetAllHtmlNodes(node, "", "id", "") {
			//fmt.Println(htmlutil.HtmlNodeToString(n))
			//debugNode(n)
			//fmt.Println(node.Attr["id"])
			for _, attr := range n.Attr {
				if attr.Key == "id" {
					internalLinks[attr.Val] = fmt.Sprintf("%s#%s", sectionFilename, attr.Val)
				}
			}
		}
	}

	//fmt.Printf("%v\n", internalLinks)

	// Make sure the last section gets added
	sections = append(sections, *section)

	e := epub.NewEpub(effectiveGoTitle)
	e.SetCover(effectiveGoCoverImg, "")

	// Iterate through each section and add it to the EPUB
	for _, section := range sections {
		sectionContent := ""
		for _, node := range section.nodes {
			nodeContent, err := htmlutil.HtmlNodeToString(&node)
			if err != nil {
				return err
			}
			sectionContent += nodeContent
		}

		_, err := e.AddSection(section.title, sectionContent, section.filename, "")
		if err != nil {
			return err
		}
	}

	err = e.Write(effectiveGoFilename)
	if err != nil {
		return err
	}

	return nil
}

func titleToFilename(title string) string {
	title = strings.ToLower(title)
	title = strings.Replace(title, " ", "-", -1)

	return fmt.Sprintf("%s.xhtml", title)
}

// TODO: remove this
func debugNode(node *html.Node) {
	fmt.Printf("type: %s\n", node.Type)
	if node.Type == html.CommentNode {
		fmt.Println("type: CommentNode")
	} else if node.Type == html.DoctypeNode {
		fmt.Println("type: DoctypeNode")
	} else if node.Type == html.DocumentNode {
		fmt.Println("type: DocumentNode")
	} else if node.Type == html.ElementNode {
		fmt.Println("type: ElementNode")
	} else if node.Type == html.ErrorNode {
		fmt.Println("type: ErrorNode")
	} else if node.Type == html.TextNode {
		fmt.Println("type: TextNode")
	}

	fmt.Printf("data: %s\n", node.Data)
	fmt.Printf("attr: %s\n", node.Attr)
	fmt.Println(htmlutil.HtmlNodeToString(node))
}

func reformatEffectiveGoFooter(footerNode *html.Node) *html.Node {
	newBrNode := func() *html.Node {
		return &html.Node{
			Type: html.ElementNode,
			Data: "br",
		}
	}
	for node := footerNode.FirstChild; node != nil; node = node.NextSibling {
		// Double all <br> elements for styling
		if node.Type == html.ElementNode && node.Data == "br" {
			footerNode.InsertBefore(newBrNode(), node)

		} else if node.Type == html.TextNode && strings.Contains(node.Data, "page") {
			node.Data = strings.Replace(node.Data, "page", "book", -1)
		}
	}
	footerNode.InsertBefore(
		newBrNode(),
		footerNode.FirstChild)
	footerNode.InsertBefore(
		newBrNode(),
		footerNode.FirstChild)
	sourceLinkNode := &html.Node{
		Type: html.ElementNode,
		Data: "a",
		Attr: []html.Attribute{
			html.Attribute{
				Key: "href",
				Val: effectiveGoUrl,
			}},
	}
	sourceLinkNode.AppendChild(&html.Node{
		Type: html.TextNode,
		Data: effectiveGoUrl,
	})
	footerNode.InsertBefore(sourceLinkNode, footerNode.FirstChild)
	footerNode.InsertBefore(
		&html.Node{
			Type: html.TextNode,
			Data: "Source: ",
		},
		footerNode.FirstChild)

	return footerNode
}
