package main

import (
	"log"
	"net/http"
	"strings"

	"golang.org/x/net/html"

	"github.com/bmaupin/go-epub"
	"github.com/bmaupin/go-util/htmlutil"
)

const (
	effectiveGoFilename  = "Effective Go.epub"
	effectiveGoSeparator = "<h2"
	effectiveGoTitle     = "Effective Go"
	effectiveGoUrl       = "https://golang.org/doc/effective_go.html"
)

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

	contentNode, err := htmlutil.GetFirstHtmlNode(doc, "div", "id", "page")

	// Remove the footer node
	htmlutil.RemoveFirstHtmlNode(contentNode, "div", "id", "footer")

	contentText, err := htmlutil.HtmlNodeToString(contentNode)

	e := epub.NewEpub(effectiveGoTitle)

	// Split up the page into sections
	for i, sectionContent := range strings.Split(contentText, effectiveGoSeparator) {
		// Ignore everything before the first separator
		if i == 0 {
			continue
		}

		// Add the separator back to the content
		sectionContent = effectiveGoSeparator + sectionContent

		// Extract the title
		sectionTitle := sectionContent[strings.Index(sectionContent, ">")+1 : strings.Index(sectionContent, "</")]

		// TODO: assign a filename to each section for internal links
		e.AddSection(sectionTitle, sectionContent, "", "")
	}

	e.Write(effectiveGoFilename)

	return nil
}
