package main

import (
	"flag"
	"log"
	"net/http"
	"os"

	"github.com/wzshiming/shorturl"
	"github.com/wzshiming/shorturl/storage/sqlite"
)

var address string
var index uint64
var storage string
var baseURL string

func init() {
	flag.StringVar(&address, "a", ":8080", "listen on the address")
	flag.Uint64Var(&index, "i", 10000, "start index")
	flag.StringVar(&storage, "s", "shorturl.db", "storage file")
	flag.StringVar(&baseURL, "b", "", "base url")
	flag.Parse()
}

func main() {
	logger := log.New(os.Stderr, "[short url] ", log.LstdFlags)

	s, err := sqlite.NewStorage(index, storage)
	if err != nil {
		logger.Println(err)
		return
	}
	h := shorturl.NewHandler(s,
		shorturl.WithBaseURL(baseURL),
	)
	logger.Printf("listen on %s", address)
	err = http.ListenAndServe(address, h)
	if err != nil {
		logger.Println(err)
	}
}
