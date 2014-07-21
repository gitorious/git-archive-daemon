package main

import (
	"net/http"
	"strings"
)

type ArchiveError int

const (
	REPOSITORY_NOT_FOUND ArchiveError = iota + 1
	REF_NOT_FOUND
)

func (e ArchiveError) Error() string {
	switch e {
	case REPOSITORY_NOT_FOUND:
		return "repository not found"
	case REF_NOT_FOUND:
		return "ref not found"
	}

	return "unknown error"
}

func (e ArchiveError) HttpCode() int {
	switch e {
	case REPOSITORY_NOT_FOUND:
		return http.StatusNotFound
	case REF_NOT_FOUND:
		return http.StatusNotFound
	}

	return http.StatusInternalServerError
}

type Server struct {
	generator interface {
		GenerateArchive(string, string, string, string) (string, error)
	}
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// TODO: respond with "method not allowed" for non-GETs

	query := r.URL.Query()

	path := strings.Trim(r.URL.Path, "/")
	if path == "" {
		path = "."
	}

	ref := query.Get("ref")
	if ref == "" {
		http.Error(w, "ref parameter is missing", http.StatusBadRequest)
		return
	}

	prefix := query.Get("prefix")

	format := query.Get("format")
	if format != "tar.gz" && format != "zip" {
		http.Error(w, "requested format is invalid", http.StatusBadRequest)
		return
	}

	filename := query.Get("filename")

	archivePath, err := s.generator.GenerateArchive(path, ref, prefix, format)

	if err != nil {
		if archiveErr, ok := err.(ArchiveError); ok {
			http.Error(w, archiveErr.Error(), archiveErr.HttpCode())
		} else {
			http.Error(w, "internal server error", http.StatusInternalServerError)
		}
		return
	}

	if filename != "" {
		w.Header().Set("Content-Disposition", "attachment; filename="+filename)
	}
	http.ServeFile(w, r, archivePath)
}
