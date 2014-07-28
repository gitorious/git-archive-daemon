package main

import (
	"errors"
	"fmt"
	"os/exec"
	"strings"
)

type Repository interface {
	Path() string
	ResolveRef(string) (string, error)
	Archive(string, string, string, string) error
}

type GitRepository struct {
	path string
}

var _ Repository = &GitRepository{}

func (r *GitRepository) Path() string {
	return r.path
}

func (r *GitRepository) ResolveRef(rev string) (string, error) {
	commit, err := r.execf("rev-parse %v", rev)

	if err != nil {
		return "", err
	}

	return strings.Trim(commit, " \n"), nil
}

func (r *GitRepository) Archive(commit, prefix, format, outPath string) error {
	if format == "zip" {
		_, err := r.execf("archive --prefix=%v --format=zip %v >%v", prefix, commit, outPath)
		return err
	} else if format == "tar.gz" {
		_, err := r.execf("archive --prefix=%v --format=tar %v | gzip >%v", prefix, commit, outPath)
		return err
	}

	return errors.New("Invalid format: " + format)
}

func (r *GitRepository) exec(gitCommand string) (string, error) {
	command := "git --git-dir=" + r.path + " " + gitCommand
	cmd := exec.Command("bash", "-o", "pipefail", "-c", command)
	output, err := cmd.Output()

	return string(output), err
}

func (r *GitRepository) execf(gitCommand string, args ...interface{}) (string, error) {
	return r.exec(fmt.Sprintf(gitCommand, args...))
}
