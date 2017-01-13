package main

import (
	"os"
	"strings"

	"github.com/crosbymichael/octokat"
	"github.com/docker/docker/pkg/term"
	"github.com/fatih/color"
)

func repo(myRepo string) octokat.Repo {
	parts := strings.SplitN(myRepo, "/", 2)
	return octokat.Repo{UserName: parts[0], Name: parts[1]}
}

func line() {
	size, err := term.GetWinsize(0)
	if err != nil {
		exitError(err)
	}

	color.New(color.FgYellow, color.Bold).Println(strings.Repeat("-", int(size.Width)))
}

func getClient() *octokat.Client {
	client := octokat.NewClient()
	client.WithToken(os.Getenv("GITHUB_TOKEN"))
	return client
}
