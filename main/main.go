package main

import (
	"net/http"

	"golang.org/x/net/context"
)

func main() {
	// TODO(dino): Move fixture creation out of test code.
	fixture, err := createFixture()
	if err != nil {
		panic(err)
	}

	ingestRequest1, err := ingestRequest(&entry{Message: "foo", TimestampMs: 5000})
	if err != nil {
		panic(err)
	}

	_, err = fixture.ingester.Ingest(context.Background(), ingestRequest1)
	if err != nil {
		panic(err)
	}

	mux := http.NewServeMux()
	fixture.mixer.RegisterHttp(mux)

	http.ListenAndServe(":12345", mux)
}
