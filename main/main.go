package main

import (
	"fmt"
	"log"
	"net"

	"dinowernli.me/almanac/appender"
	"dinowernli.me/almanac/discovery"
	"dinowernli.me/almanac/index"
	"dinowernli.me/almanac/mixer"
	pb_logging "dinowernli.me/almanac/proto"
	"dinowernli.me/almanac/storage"

	"github.com/blevesearch/bleve"
	"google.golang.org/grpc"
)

const (
	maxEntriesPerChunk = 4
)

type data struct {
	Name string
}

func main() {
	memoryStorage := storage.NewInMemoryStorage()

	err := memoryStorage.Write("fileid", []byte("hello"))
	if err != nil {
		log.Fatalf("unable to write to storage: %v", err)
	}

	index1, err := index.NewIndex()
	if err != nil {
		log.Fatalf("failed to create index service: %v", err)
	}
	index1.Index("id1", &data{Name: "foo"})
	index1.Index("id2", &data{Name: "bar"})
	log.Println("created index")

	server := grpc.NewServer()

	// Our index happens to implement the Mixer service, register it.
	pb_logging.RegisterMixerServer(server, index1)

	nextAppenderId := 0
	appender, err := appender.New(fmt.Sprintf("appender%d", nextAppenderId), memoryStorage, maxEntriesPerChunk)
	nextAppenderId++
	if err != nil {
		log.Fatalf("failed to create appender: %v", err)
	}
	pb_logging.RegisterAppenderServer(server, appender)

	listen, err := net.Listen("tcp", fmt.Sprintf(":%v", 12345))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	go func() {
		server.Serve(listen)
	}()
	log.Println("started grpc server")

	d, err := discovery.New([]string{"localhost:12345"})
	if err != nil {
		log.Fatalf("failed to start discovery: %v", err)
	}
	_ = mixer.New(memoryStorage, d)

	conn, err := grpc.Dial("localhost:12345", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("failed to dial: %v", err)
	}

	mixerClient := pb_logging.NewMixerClient(conn)
	indexAlias := bleve.NewIndexAlias(index.NewRemoteIndex(mixerClient))
	bleveQuery := bleve.NewMatchQuery("foo")
	bleveSearch := bleve.NewSearchRequest(bleveQuery)
	bleveResult, err := indexAlias.Search(bleveSearch)
	if err != nil {
		log.Fatalf("search failed: %v", err)
	}

	log.Printf("done searching remotely: %v", bleveResult)
}
