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
	effectiveGoCoverImg   = "covers/effective-go.png"
	effectiveGoFilename   = "Effective Go.epub"
	effectiveGoSectionTag = "h2"
	effectiveGoSeparator  = "<h2"
	effectiveGoTitle      = "Effective Go"
	effectiveGoUrl        = "https://golang.org/doc/effective_go.html"
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

	// Remove the footer node
	// TODO: add this to the title page
	htmlutil.RemoveFirstHtmlNode(containerNode, "div", "id", "footer")

	sections := []epubSection{}
	// TODO: add a title to the first section?
	// TODO: set filename?
	// TODO: remove the <div id="nav"></div> element from the first section
	section := &epubSection{}
	//	internalLinks := make(map[string]string)

	// Iterate through each child node
	for n := containerNode.FirstChild; n != nil; n = n.NextSibling {
		// If we find the section tag
		if n.Type == html.ElementNode && n.Data == effectiveGoSectionTag {
			// Add the previous section to the slice of sections
			sections = append(sections, *section)

			sectionTitle := n.FirstChild.Data

			// Start a new section
			section = &epubSection{
				filename: titleToFilename(sectionTitle),
				title:    sectionTitle,
			}
		}

		// Append the current node to the current section
		section.nodes = append(section.nodes, *n)

		/*
					for _, node := range GetHtmlNodes(containerNode, "", "id", "", -1) {
				fmt.Println(htmlutil.HtmlNodeToString(node))
			}
		*/
	}

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

func debugNode(n *html.Node) {
	fmt.Printf("type: %s\n", n.Type)
	if n.Type == html.CommentNode {
		fmt.Println("type: CommentNode")
	} else if n.Type == html.DoctypeNode {
		fmt.Println("type: DoctypeNode")
	} else if n.Type == html.DocumentNode {
		fmt.Println("type: DocumentNode")
	} else if n.Type == html.ElementNode {
		fmt.Println("type: ElementNode")
	} else if n.Type == html.ErrorNode {
		fmt.Println("type: ErrorNode")
	} else if n.Type == html.TextNode {
		fmt.Println("type: TextNode")
	}

	fmt.Printf("data: %s\n", n.Data)
	fmt.Println(htmlutil.HtmlNodeToString(n))
}
