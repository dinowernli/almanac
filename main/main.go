package main

import (
	"github.com/blevesearch/bleve"
	"io/ioutil"
	"log"
)

type data struct {
	Name string
}

func main() {
	dir, err := ioutil.TempDir("", "index.bleve")
	if err != nil {
		log.Fatalf("failed to create tempfile: %v", err)
	}
	log.Printf("created directory: %s", dir)

	mapping := bleve.NewIndexMapping()
	index, err := bleve.New(dir, mapping)
	if err != nil {
		log.Fatalf("failed to create index: %v", err)
	}

	index.Index("id1", &data{Name: "foo"})
	index.Index("id2", &data{Name: "bar"})

	query := bleve.NewMatchQuery("foo")
	search := bleve.NewSearchRequest(query)
	searchResults, err := index.Search(search)
	if err != nil {
		log.Fatalf("failed to search: %v", err)
	}

	log.Println(searchResults)
}
