package main

import (
	"context"
	"fmt"
	"os"

	"github.com/fatih/color"
	"github.com/google/go-github/github"
	"github.com/urfave/cli"
)

func listIssue(ctx *cli.Context) {
	client := getClient()

	owner, repo, err := repo()
	if err != nil {
		exitError(err)
	}

	newIssues := []*github.Issue{}

	for page := 1; page < ctx.Int("max-pages"); page++ {
		params := &github.IssueListByRepoOptions{
			State:     ctx.String("state"),
			Sort:      ctx.String("sort-by"),
			Direction: ctx.String("direction"),
			ListOptions: github.ListOptions{
				Page: page,
			},
		}

		prs, _, err := client.Issues.ListByRepo(context.Background(), owner, repo, params)
		if err != nil {
			exitError(err)
		}

		if len(prs) == 0 {
			break
		}

		newIssues = append(newIssues, prs...)
	}

	for _, issue := range newIssues {
		color.New(color.FgWhite).Printf("[ %d ] ", issue.GetNumber())
		color.New(color.FgBlue).Printf("(%s) ", issue.User.GetLogin())
		fmt.Fprintf(os.Stdout, "%s\n", issue.GetTitle())
	}
}
