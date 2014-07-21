package main

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type testArchiveGenerator struct {
	err error
}

func (g testArchiveGenerator) GenerateArchive(path, ref, prefix, format string) (string, error) {
	if g.err != nil {
		return "", g.err
	}

	filename := fmt.Sprintf("%v-%v-%v-%v", path, ref, prefix, format)
	filename = strings.Replace(filename, "/", "-", -1)
	filename = strings.Replace(filename, ".", "-", -1)

	return "fixtures/files/" + filename, nil
}

func TestServer_ServeHTTP(t *testing.T) {
	var tests = []struct {
		requestPath        string
		generator          testArchiveGenerator
		expectedHTTPStatus int
		expectedBody       string
	}{
		{ // all good, zip
			"foo?ref=master&format=zip&prefix=prefix",
			testArchiveGenerator{},
			200,
			"foo-master-prefix-zip\n",
		},
		{ // all good, tar.gz
			"bar?ref=fixes&format=tar.gz&prefix=prefix",
			testArchiveGenerator{},
			200,
			"bar-fixes-prefix-tar-gz\n",
		},
		{ // missing format
			"foo?ref=blah&prefix=prefix",
			testArchiveGenerator{},
			400,
			"requested format is invalid\n",
		},
		{ // invalid format
			"foo?ref=blah&format=rar&prefix=prefix",
			testArchiveGenerator{},
			400,
			"requested format is invalid\n",
		},
		{ // missing ref
			"foo?format=zip&prefix=prefix",
			testArchiveGenerator{},
			400,
			"ref parameter is missing\n",
		},
		{ // missing prefix, but that's ok
			"foo?ref=next&format=zip",
			testArchiveGenerator{},
			200,
			"foo-next--zip\n",
		},
		{ // empty path, but that's ok
			"?ref=next&format=zip&prefix=prefix",
			testArchiveGenerator{},
			200,
			"--next-prefix-zip\n",
		},
		{ // wrong repo path
			"qux?ref=master&format=zip",
			testArchiveGenerator{REPOSITORY_NOT_FOUND},
			404,
			"repository not found\n",
		},
		{ // invalid ref
			"foo?ref=blah&format=zip",
			testArchiveGenerator{REF_NOT_FOUND},
			404,
			"ref not found\n",
		},
		{ // other error
			"foo?ref=master&format=zip&prefix=prefix",
			testArchiveGenerator{errors.New("unknown error")},
			500,
			"internal server error\n",
		},
	}

	for _, test := range tests {
		server := Server{test.generator}

		r, _ := http.NewRequest("GET", "http://localhost/"+test.requestPath, nil)
		w := httptest.NewRecorder()

		server.ServeHTTP(w, r)

		body := w.Body.String()

		if w.Code != test.expectedHTTPStatus {
			t.Errorf("expected status %v, got %v for %v", test.expectedHTTPStatus, w.Code, test)
		}

		if body != test.expectedBody {
			t.Errorf("expected body \"%v\", got \"%v\" for %v", test.expectedBody, body, test)
		}
	}
}
