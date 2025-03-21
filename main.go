package main

import (
	"fmt"
	"log"
	"log/slog"
	"os"
	"strings"
	"time"

	git "github.com/go-git/go-git/v5"
	gitconfig "github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
)

func init() {
	slog.SetDefault(
		slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelDebug,
		})),
	)
}

func main() {
	if len(os.Args) < 2 {
		log.Fatal("Usage: create-tag <tag>")
	}
	tagName := os.Args[1]

	// Open the Git repository
	repo, err := git.PlainOpen(".")
	if err != nil {
		log.Fatal("Error opening repository:", err)
	}

	// Get the HEAD reference
	headRef, err := repo.Head()
	if err != nil {
		log.Fatal("Error getting HEAD reference:", err)
	}

	// Get the commit history
	commitIter, err := repo.Log(&git.LogOptions{From: headRef.Hash()})
	if err != nil {
		log.Fatal("Error getting commit log:", err)
	}

	// Get global config
	config, err := repo.ConfigScoped(gitconfig.GlobalScope)
	if err != nil {
		log.Fatal("Error fetching global config:", err)
	}

	// Find the latest tag
	lastTag, err := getLatestTag(repo)
	if err != nil {
		log.Println("No previous tags found, using full history.")
	}

	var commits []string
	if err := commitIter.ForEach(func(c *object.Commit) error {
		if lastTag != "" && c.Hash.String() == lastTag {
			return nil
		}
		commits = append(commits, " - "+c.Message)
		return nil
	}); err != nil {
		log.Fatal("Error iterating commits:", err)
	}

	if len(commits) == 0 {
		log.Fatal("No new commits to tag.")
	}

	// Create the tag
	tagMessage := tagName + "\n\n" + strings.Join(commits, "\n")
	if _, err := repo.CreateTag(tagName, headRef.Hash(), &git.CreateTagOptions{
		Message: tagMessage,
		Tagger: &object.Signature{
			Name:  config.User.Name,
			Email: config.User.Email,
			When:  time.Now(),
		},
	}); err != nil {
		log.Fatal("Error creating tag:", err)
	}

	fmt.Println("Tag", tagName, "created with message:\n\n", tagMessage)
}

// getLatestTag finds the latest Git tag
func getLatestTag(repo *git.Repository) (string, error) {
	tags, err := repo.Tags()
	if err != nil {
		return "", err
	}

	var latestTag string
	err = tags.ForEach(func(ref *plumbing.Reference) error {
		latestTag = ref.Hash().String()
		return nil
	})
	if err != nil {
		return "", err
	}

	return latestTag, nil
}
