package main

import (
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"regexp"
	"strings"
	"text/template"
	"time"
)

type Feed struct {
	XMLName xml.Name    `xml:"feed"`
	Entries []FeedEntry `xml:"entry"`
}
type Entry struct {
	PubDate time.Time
	Title   string
	Content string
	Tags    []string
	Draft   bool
}

type FeedEntry struct {
	PubDate    time.Time  `xml:"published"`
	Categories []Category `xml:"category"`
	Title      string     `xml:"title"`
	Content    string     `xml:"content"`
	Control    Control    `xml:"control"`
}

type Control struct {
	XMLName xml.Name
	Draft   string `xml:"draft"`
}

type Category struct {
	Scheme string `xml:"scheme,attr"`
	Term   string `xml:"term,attr"`
}

type XMLTime struct {
	time.Time
}

func (t *XMLTime) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	var v string
	d.DecodeElement(&v, &start)
	//2007-04-17T15:09:01.730-07:0
	parse, err := time.Parse(time.RFC3339, v)
	if err != nil {
		return err
	}
	*t = XMLTime{parse}
	return nil
}

func fromFeed(feed Feed) []Entry {
	var entries []Entry
	const TAG_SCHEME = "http://www.blogger.com/atom/ns#"
	const KIND_SCHEME = "http://schemas.google.com/g/2005#kind"
	for _, entry := range feed.Entries {
		var tags []string
		found := false
		for _, cat := range entry.Categories {
			if cat.Scheme == TAG_SCHEME {
				tags = append(tags, cat.Term)
			} else if cat.Scheme == KIND_SCHEME && strings.HasSuffix(cat.Term, "kind#post") {
				found = true
			}
		}
		if found {
			entries = append(entries, Entry{
				PubDate: entry.PubDate,
				Title:   entry.Title,
				Content: entry.Content,
				Tags:    tags,
				Draft:   entry.Control.Draft == "yes",
			})
		}
	}
	return entries
}

const EntryTemplate = `---
title: "{{.Title}}"
date: {{ date .PubDate }}
draft: true
tags: [ {{range $i, $t := .Tags}}{{if $i}}, {{end}}"{{.}}"{{end}} ]
---
{{.Content}}`

func sanitize(s string) string {
	re := regexp.MustCompile("[^A-Za-z0-9-]+")
	base := strings.ReplaceAll(strings.ToLower(s), " ", "-")
	return re.ReplaceAllLiteralString(base, "")
}
func main() {
	f, err := os.Open("out.xml")
	if err != nil {
		log.Fatal("Failed to open file", err)
	}
	defer f.Close()
	buf, err := ioutil.ReadAll(f)
	if err != nil {
		log.Fatal("Failed to read file", err)
	}
	var feed Feed
	err = xml.Unmarshal(buf, &feed)
	//fmt.Printf("%+v", feed)
	entries := fromFeed(feed)
	fmt.Printf("Found %d posts\n", len(entries))
	err = os.MkdirAll("posts", os.ModeDir|0755)

	if err != nil {
		log.Fatal("Failed to create directory", err)
	}
	funcMap := template.FuncMap{
		"date": func(t time.Time) string {
			return t.Format(time.RFC3339)
		},
	}
	tmpl := template.Must(template.New("entry").Funcs(funcMap).Parse(EntryTemplate))
	for _, entry := range entries {
		out, err := os.Create("posts/" + sanitize(entry.Title) + ".md")
		if err != nil {
			log.Fatal("Failed to create file", err)
		}
		defer out.Close()
		err = tmpl.Execute(out, entry)
		if err != nil {
			log.Println("Failed to merge template", err)
		}
		fmt.Printf("%s %s %t\n", entry.Title, entry.Tags, entry.Draft)
	}

}
