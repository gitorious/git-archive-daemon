package main

import (
	"flag"
	"log"
	"net/http"
	"os"
)

func main() {
	var (
		numWorkers = flag.Int("w", 10, "Number of workers")
		tmpDir     = flag.String("t", os.TempDir(), "Tmp dir for archive generation")
		cacheDir   = flag.String("c", ".", "Cache dir for storing archives")
		addr       = flag.String("l", ":5000", "Address/port to listen on")
	)
	flag.Parse()

	jobs := make(chan *ArchiveJob)
	results := make(chan *ArchiveJob)

	for n := 0; n < *numWorkers; n++ {
		go ArchiveWorker(jobs, results, *tmpDir)
	}

	jobs, results = ArchiveCache(jobs, results, *cacheDir)

	requests := make(chan *ArchiveRequest)
	go RequestMux(requests, jobs, results)

	repositoryStore := &GitRepositoryStore{"repositories"}
	archiveGenerator := &ArchiveGenerator{repositoryStore, requests}

	server := &Server{archiveGenerator}

	log.Fatal(http.ListenAndServe(*addr, server))
}
