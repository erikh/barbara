package main

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"

	"github.com/fatih/color"
	"github.com/google/go-github/github"
	"github.com/urfave/cli"
)

func getIssue(ctx *cli.Context) {
	if len(ctx.Args()) != 1 {
		exitError(errors.New("invalid parameters"))
	}

	client := getClient()

	owner, repo, err := repo()
	if err != nil {
		exitError(err)
	}

	num, err := strconv.Atoi(ctx.Args()[0])
	if err != nil {
		exitError(err)
	}

	issue, _, err := client.Issues.Get(context.Background(), owner, repo, num)
	if err != nil {
		exitError(err)
	}

	allComments := []*github.IssueComment{}
	var i int

	for {
		i++
		comments, _, err := client.Issues.ListComments(context.Background(), owner, repo, num, &github.IssueListCommentsOptions{
			ListOptions: github.ListOptions{
				Page: i,
			},
		})
		if err != nil {
			exitError(err)
		}

		if len(comments) == 0 {
			break
		}

		allComments = append(allComments, comments...)
	}

	color.Output = os.Stdout

	line()
	color.New(color.FgHiBlue).Printf("From: %s\n", issue.User.GetLogin())
	color.New(color.FgHiBlue).Printf("Last Updated: %v\n", issue.GetUpdatedAt())
	color.New(color.FgHiBlue).Printf("Title: %s\n", issue.GetTitle())
	color.New(color.FgHiBlue).Printf("Number: %d\n", issue.GetNumber())
	color.New(color.FgHiBlue).Printf("URL: %s\n", issue.GetHTMLURL())

	stateColor := color.New()
	switch issue.GetState() {
	case "open":
		stateColor = color.New(color.FgGreen)
	case "closed":
		stateColor = color.New(color.FgRed)
	}

	stateColor.Printf("State: %s\n", issue.GetState())

	line()
	fmt.Println(issue.GetBody())

	for _, comment := range allComments {
		fmt.Println()
		line()
		color.New(color.FgWhite).Printf("From: %s\n", comment.User.GetLogin())
		color.New(color.FgWhite).Printf("Date: %s\n", comment.CreatedAt.Local())
		line()
		fmt.Println()
		fmt.Println(comment.GetBody())
	}

	fmt.Println()

}

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

func replyIssue(ctx *cli.Context) {
	args := ctx.Args()

	if len(args) != 1 {
		exitError(errors.New("invalid arguments"))
	}

	client := getClient()

	owner, repo, err := repo()
	if err != nil {
		exitError(err)
	}

	num, err := strconv.Atoi(args[0])
	if err != nil {
		exitError(err)
	}

	f, err := ioutil.TempFile("", "barb-")
	if err != nil {
		exitError(err)
	}
	f.Close()
	defer os.Remove(f.Name())

	if err := runProgram(os.Getenv("EDITOR"), f.Name()); err != nil {
		exitError(err)
	}

	body, err := ioutil.ReadFile(f.Name())
	if err != nil {
		exitError(err)
	}

	_, _, err = client.Issues.CreateComment(context.Background(), owner, repo, num, &github.IssueComment{Body: github.String(string(body))})
	if err != nil {
		exitError(err)
	}

	fmt.Println("Comment posted!")
}

func reopenIssue(ctx *cli.Context) {
	editIssue(ctx, "opened")
}

func closeIssue(ctx *cli.Context) {
	editIssue(ctx, "closed")
}

func editIssue(ctx *cli.Context, state string) {
	args := ctx.Args()

	if len(args) != 1 {
		exitError(errors.New("invalid arguments"))
	}

	client := getClient()

	num, err := strconv.Atoi(args[0])
	if err != nil {
		exitError(err)
	}

	owner, repo, err := repo()
	if err != nil {
		exitError(err)
	}

	_, _, err = client.Issues.Edit(context.Background(), owner, repo, num, &github.IssueRequest{
		State: github.String(state),
	})

	if err != nil {
		exitError(err)
	}

	fmt.Println("Issue", num, state+"!")
}
