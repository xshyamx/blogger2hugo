package main

import (
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"strings"
	"os"
	"time"
	"log"
)

type Feed struct {
	XMLName xml.Name `xml:"feed"`
	Entries []Entry `xml:"entry"`
}

type Entry struct {
	PubDate XMLTime `xml:"published"`
	Category struct {
		Kind Kind `xml:"term,attr"`
	} `xml:"category"`
	Title string `xml:"title"`
	Content string `xml:"content"`
}

type XMLTime struct {
	time.Time
}

type Kind struct {
	string
}

func (k *Kind) UnmarshalXMLAttr(attr xml.Attr) error {
  val := attr.Value
	i := strings.LastIndex(val, "#")
	if i != -1 {
		fmt.Println("UnmarshalXMLAttr", val, val[i+1:])
		*k = Kind{val[i+1:]}
		fmt.Println("kind", *k)
	}
	return nil
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
	for _, entry := range feed.Entries {
		fmt.Printf("%s %s\n", entry.Title, entry.Category.Kind)
	}
}
