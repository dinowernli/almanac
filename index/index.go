package index

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/blevesearch/bleve"

	"golang.org/x/net/context"
)

// Index wraps a bleve index and presents a serializable interface.
type Index struct {
	index bleve.Index
	path  string
}

// openIndex returns an index backed by the contents on disk at the specified
// path. The caller responsible for eventually calling Close() on the returned
// instance to release disk resources.
func openIndex(dir string) (*Index, error) {
	index, err := bleve.Open(dir)
	if err != nil {
		return nil, fmt.Errorf("failed to create index: %v", err)
	}
	return &Index{index: index, path: dir}, nil
}

// NewIndex returns an instance of index backed by an temporary location on disk.
// The caller is responsible for eventually calling Close() on the returned index.
func NewIndex() (*Index, error) {
	dir, err := ioutil.TempDir("", "index.bleve")
	if err != nil {
		return nil, fmt.Errorf("failed to create tempfile: %v", err)
	}

	mapping := bleve.NewIndexMapping()
	index, err := bleve.New(dir, mapping)
	if err != nil {
		return nil, fmt.Errorf("failed to create index: %v", err)
	}
	return &Index{index: index, path: dir}, nil
}

// Search executes a search on the index and returns the ids of the log
// entries which match the search.
func (i *Index) Search(ctx context.Context, query string, num int32) ([]string, error) {
	request := bleve.NewSearchRequestOptions(
		bleve.NewMatchQuery(query),
		int(num),
		0,     // from
		false) // explain

	response, err := i.index.Search(request)
	if err != nil {
		return nil, fmt.Errorf("unable to search index: %v", err)
	}

	result := []string{}
	for _, hit := range response.Hits {
		result = append(result, hit.ID)
	}
	return result, nil
}

func (i *Index) Bleve() bleve.Index {
	// TODO(dino): Implementation detail, teach this to accept bleve
	// searches instead.
	return i.index
}

func (i *Index) Index(id string, data interface{}) error {
	return i.index.Index(id, data)
}

// Close releases any resources held by this instance. No other methods must
// be called after this.
func (i *Index) Close() error {
	return os.RemoveAll(i.path)
}
