package main

import (
	"archive/zip"
	"bing-scraping/metadata"
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"

	"github.com/PuerkitoBio/goquery"
)

func handler(i int, s *goquery.Selection) {
	url, ok := s.Find("a").Attr("href")
	if !ok {
		return
	}

	fmt.Printf("%d: %s\n", i, url)
	res, err := http.Get(url)
	if err != nil {
		return
	}
	buf, err := io.ReadAll(res.Body)
	if err != nil {
		return
	}
	defer res.Body.Close()

	r, err := zip.NewReader(bytes.NewReader(buf), int64(len(buf)))
	if err != nil {
		return
	}

	cp, ap, err := metadata.NewProperties(r)
	if err != nil {
		return
	}

	log.Printf(
		"%25s %25s - %s %s\n",
		cp.Creator,
		cp.LastModifiedBy,
		ap.Application,
		ap.GetMajorVersion(),
	)
}

func main() {
	if len(os.Args) != 3 {
		log.Fatalf("Missing required argument.\n\tUsage: main.go domain ext\n")
	}
	domain := os.Args[1]
	filetype := os.Args[2]

	q := fmt.Sprintf(
		"site:%s && filetype:%s && instreamset:(url title):%s",
		domain,
		filetype,
		filetype,
	)

	search := fmt.Sprintf("https://www.bing.com/search?q=%s", url.QueryEscape(q))
	client := &http.Client{}
	req, err := http.NewRequest("GET", search, nil)
	if err != nil {
		log.Fatalln("Failed while creating request")
	}

	// We need to spoof user agent or else Bing will not return any results.
	req.Header.Add("User-Agent", `Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/39.0.2171.27 Safari/537.36`)
	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf("Failed to get documents with error: %s", err.Error())
	}
	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)

	if err != nil {
		log.Panicln(err)
	}
	s := "html body div#b_content ol#b_results li.b_algo h2"
	doc.Find(s).Each(handler)
}
