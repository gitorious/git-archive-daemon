package main

import (
	"flag"
	"log"
	"net/http"
	"os"
)

func main() {
	var (
		reposDir   = flag.String("r", ".", "Directory containing git repositories")
		cacheDir   = flag.String("c", ".", "Cache dir for storing archives")
		tmpDir     = flag.String("t", os.TempDir(), "Tmp dir for archive generation")
		addr       = flag.String("l", ":5000", "Address/port to listen on")
		numWorkers = flag.Int("w", 10, "Number of workers")
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

	repositoryStore := &GitRepositoryStore{*reposDir}
	archiveGenerator := &ArchiveGenerator{repositoryStore, requests}

	server := &Server{archiveGenerator}

	log.Fatal(http.ListenAndServe(*addr, server))
}
