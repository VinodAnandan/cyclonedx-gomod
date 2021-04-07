package gomod

import (
	"fmt"
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/storer"

	"github.com/CycloneDX/cyclonedx-gomod/internal/util"
)

// GetModuleVersion attempts to detect a given module's version by first
// calling GetVersionFromTag and if that fails, GetPseudoVersion on it.
func GetModuleVersion(modulePath string) (string, error) {
	if tagVersion, err := GetVersionFromTag(modulePath); err != nil {
		// TODO: Log err in DEBUG / verbose level
		pseudoVersion, err := GetPseudoVersion(modulePath)
		if err != nil {
			return "", err
		}
		return pseudoVersion, nil
	} else {
		return tagVersion, nil
	}
}

// GetPseudoVersion constructs a pseudo version for a Go module at a given path.
// Note that this is only possible when path points to a Git repository.
// See https://golang.org/ref/mod#pseudo-versions
func GetPseudoVersion(modulePath string) (string, error) {
	repo, err := git.PlainOpen(modulePath)
	if err != nil {
		return "", err
	}

	headRef, err := repo.Head()
	if err != nil {
		return "", err
	}

	headCommit, err := repo.CommitObject(headRef.Hash())
	if err != nil {
		return "", err
	}

	commitHash := headCommit.Hash.String()[:12]
	commitDate := headCommit.Author.When.Format("20060102150405")

	return fmt.Sprintf("v0.0.0-%s-%s", commitDate, commitHash), nil
}

// GetVersionFromTag checks if the current commit is annotated with a tag and if yes, returns that tag's name.
// Note that this is only possible when path points to a Git repository.
func GetVersionFromTag(modulePath string) (string, error) {
	repo, err := git.PlainOpen(modulePath)
	if err != nil {
		return "", err
	}

	headRef, err := repo.Head()
	if err != nil {
		return "", err
	}

	tags, err := repo.Tags()
	if err != nil {
		return "", err
	}

	tagName := ""
	err = tags.ForEach(func(reference *plumbing.Reference) error {
		if reference.Hash() == headRef.Hash() && util.StartsWith(reference.Name().String(), "refs/tags/v") {
			tagName = strings.TrimPrefix(reference.Name().String(), "refs/tags/")
			return storer.ErrStop // break
		}
		return nil
	})
	if err != nil {
		return "", err
	}

	if tagName == "" {
		return "", plumbing.ErrObjectNotFound
	}

	return tagName, nil
}