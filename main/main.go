package main

import (
	"fmt"
	"github.com/blevesearch/bleve/mapping"
	"github.com/blevesearch/bleve/search"
)

func main() {
	_ = mapping.NewIndexMapping()

	var s *search.Location
	s = nil
	fmt.Printf("hello world: %s\n", s)
}
