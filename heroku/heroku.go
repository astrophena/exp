// Package heroku provides functionality for deploying Heroku apps
// (WIP).
package heroku // import "go.astrophena.name/exp/heroku"

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

// Deploy deploys an app with revision ref from dir to Heroku,
// authenticating with token.
func Deploy(app, ref, dir, token string) error {
	// Create a source tarball.
	var buf bytes.Buffer
	if err := createArchive(dir, &buf); err != nil {
		return fmt.Errorf("unable to create archive: %w", err)
	}

	// Upload a source tarball.
	sr := &sourcesResponse{}
	if err := call(
		http.MethodPost,
		fmt.Sprintf("apps/%s/sources", app),
		token,
		nil,
		&sr,
	); err != nil {
		return err
	}

	if err := uploadSourceTarball(sr.SourceBlob.PutURL, &buf); err != nil {
		return err
	}

	// Create a build.
	br := &buildRequest{}
	br.SourceBlob.URL = sr.SourceBlob.GetURL
	br.SourceBlob.Version = ref
	brr := &buildResponse{}
	if err := call(
		http.MethodPost,
		fmt.Sprintf("apps/%s/builds", app),
		token,
		br,
		brr,
	); err != nil {
		return err
	}

	return nil
}

type sourcesResponse struct {
	SourceBlob struct {
		GetURL string `json:"get_url"`
		PutURL string `json:"put_url"`
	} `json:"source_blob"`
}

type buildResponse struct {
	CreatedAt  string `json:"created_at"`
	ID         string `json:"id"`
	SourceBlob struct {
		URL     string `json:"url"`
		Version string `json:"version"`
	} `json:"source_blob"`
	OutputStreamURL string `json:"output_stream_url"`
	Status          string `json:"status"`
	UpdatedAt       string `json:"updated_at"`
	User            struct {
		Email string `json:"email"`
		ID    string `json:"id"`
	} `json:"user"`
}

type buildRequest struct {
	SourceBlob struct {
		URL     string `json:"url"`
		Version string `json:"version"`
	} `json:"source_blob"`
}

var httpClient = &http.Client{Timeout: 10 * time.Second}

func uploadSourceTarball(url string, r io.Reader) error {
	req, err := http.NewRequest(http.MethodPut, url, r)
	if err != nil {
		return err
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}

func call(method, path, token string, req, resp interface{}) error {
	var reqr io.Reader
	if req != nil {
		var buf bytes.Buffer

		enc := json.NewEncoder(&buf)
		if err := enc.Encode(req); err != nil {
			return err
		}

		reqr = &buf
	}

	r, err := http.NewRequest(
		method,
		fmt.Sprintf("https://api.heroku.com/%s", path),
		reqr,
	)
	if err != nil {
		return fmt.Errorf("unable to create request to %s: %w", path, err)
	}

	r.Header.Add("Accept", "application/vnd.heroku+json; version=3")
	r.Header.Add("Content-Type", "application/json; charset=utf-8")
	r.Header.Add("Authorization", "Bearer "+token)

	rr, err := httpClient.Do(r)
	if err != nil {
		return fmt.Errorf("unable to call %s: %w", path, err)
	}

	dec := json.NewDecoder(rr.Body)
	if err := dec.Decode(resp); err != nil {
		return fmt.Errorf("unable to decode response of %s: %w", path, err)
	}
	defer rr.Body.Close()

	return nil
}

func createArchive(dir string, wr io.Writer) error {
	// Determine which files to include in the archive.
	fis, err := ioutil.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("unable to read directory %s: %w", dir, err)
	}

	// Create new writers for gzip and tar.
	//
	// These writers are chained. Writing to the tar writer will write
	// to the gzip writer, which in turn will write to the "wr" writer.

	gw := gzip.NewWriter(wr)
	defer gw.Close()

	tw := tar.NewWriter(gw)
	defer tw.Close()

	for _, fi := range fis {
		err := addToArchive(tw, dir, fi.Name())
		if err != nil {
			return fmt.Errorf("unable to add to archive: %w", err)
		}
	}

	return nil
}

func addToArchive(tw *tar.Writer, dir, file string) error {
	// Open the file which will be written into the archive.
	f, err := os.Open(filepath.Join(dir, file))
	if err != nil {
		return err
	}
	defer f.Close()

	// Get file info about our file providing file size, mode, etc.
	fi, err := f.Stat()
	if err != nil {
		return err
	}

	// Create a tar header from the file info data.
	header, err := tar.FileInfoHeader(fi, fi.Name())
	if err != nil {
		return err
	}

	// Use full path as name (FileInfoHeader only takes the basename).
	//
	// If we don't do this the directory structure would not be
	// preserved.
	header.Name = file

	// Write file header to the tar archive.
	if err := tw.WriteHeader(header); err != nil {
		return err
	}

	// Copy file content to tar archive.
	if _, err := io.Copy(tw, f); err != nil {
		return err
	}

	return nil
}
