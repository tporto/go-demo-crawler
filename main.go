package main

import (
	"flag"
	"fmt"
	"go-crawler/db"
	"net/http"
	"net/url"
	"time"

	"golang.org/x/net/html"
)

type VisitedLink struct {
	Website     string    `bson:"website"`
	Link        string    `bson:"link"`
	VisitedDate time.Time `bson:"visited_date"`
}

var link string

func init() {
	flag.StringVar(&link, "url", "https://aprendagolang.com.br", "url para iniciar as visitas")
}

func main() {
	flag.Parse()

	done := make(chan bool)
	go visitLink(link)

	<-done
}

func visitLink(link string) {
	fmt.Printf("visitando link %s...\n", link)

	resp, err := http.Get(link)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Printf("erro com o status code: %d\n", resp.StatusCode)
	}

	doc, err := html.Parse(resp.Body)
	if err != nil {
		panic(err)
	}

	extractLinks(doc)
}

func extractLinks(node *html.Node) {
	if node.Type == html.ElementNode && node.Data == "a" {
		for _, attr := range node.Attr {
			if attr.Key != "href" {
				continue
			}

			link, err := url.Parse(attr.Val)
			if err != nil || link.Scheme == "" || link.Scheme == "mailto" {
				continue
			}

			if db.VisitedLink(link.String()) {
				fmt.Printf("link já visitado: %s\n", link)
				continue
			}

			visitedLink := VisitedLink{
				Website:     link.Host,
				Link:        link.String(),
				VisitedDate: time.Now(),
			}

			db.Insert("links", visitedLink)

			go visitLink(link.String())
		}
	}

	for c := node.FirstChild; c != nil; c = c.NextSibling {
		extractLinks(c)
	}
}
