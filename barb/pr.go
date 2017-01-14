package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/crosbymichael/octokat"
	"github.com/fatih/color"
	"github.com/urfave/cli"
)

func getPRs(client *octokat.Client, repo octokat.Repo, state, direction, sortBy string) ([]*octokat.PullRequest, error) {
	prs, err := client.PullRequests(repo, &octokat.Options{QueryParams: map[string]string{"state": state, "direction": direction, "sort": sortBy}})
	if err != nil {
		return nil, err
	}

	newPulls := []*octokat.PullRequest{}

	for _, pull := range prs {
		newPulls = append(newPulls, pull)
	}

	return newPulls, nil
}

func mergePR(ctx *cli.Context) {
	client := getClient()
	args := ctx.Args()
	if len(args) != 1 {
		exitError(errors.New("invalid arguments"))
	}

	myRepo, err := repo()
	if err != nil {
		exitError(err)
	}

	_, err = client.MergePullRequest(myRepo, args[0], nil)
	if err != nil {
		exitError(err)
	}

	fmt.Printf("PR #%s successfully merged!\n", args[0])
}

func createPR(ctx *cli.Context) {
	client := getClient()

	args := ctx.Args()

	if len(args) != 1 || ctx.String("title") == "" {
		exitError(errors.New("invalid arguments"))
	}

	myRepo, err := repo()
	if err != nil {
		exitError(err)
	}

	f, err := ioutil.TempFile("", "barbara-edit")
	if err != nil {
		exitError(err)
	}
	f.Close()

	if err := runProgram(os.Getenv("EDITOR"), f.Name()); err != nil {
		exitError(err)
	}

	content, err := ioutil.ReadFile(f.Name())
	if err != nil {
		exitError(err)
	}

	fmt.Println(args[0])

	pr, err := client.CreatePullRequest(myRepo, &octokat.Options{
		Params: map[string]string{
			"title": ctx.String("title"),
			"body":  string(content),
			"base":  ctx.String("base"),
			"head":  args[0],
		},
	})

	if err != nil {
		exitError(err)
	}

	fmt.Printf("PR %d created!\n", pr.Number)
}

func replyPR(ctx *cli.Context) {
	client := getClient()
	if len(ctx.Args()) != 1 {
		exitError(errors.New("invalid arguments"))
	}

	args := ctx.Args()
	myRepo, err := repo()
	if err != nil {
		exitError(err)
	}

	f, err := ioutil.TempFile("", "barbara-edit")
	if err != nil {
		exitError(err)
	}
	f.Close()
	defer os.Remove(f.Name())

	if err := runProgram(os.Getenv("EDITOR"), f.Name()); err != nil {
		exitError(err)
	}

	content, err := ioutil.ReadFile(f.Name())
	if err != nil {
		exitError(err)
	}

	_, err = client.AddComment(myRepo, args[0], string(content))
	if err != nil {
		exitError(err)
	}

	fmt.Printf("Comment on ticket %s posted!\n", args[0])
}

func singlePR(ctx *cli.Context) {
	client := getClient()

	if len(ctx.Args()) != 1 {
		exitError(errors.New("invalid arguments"))
	}

	args := ctx.Args()

	myRepo, err := repo()
	if err != nil {
		exitError(err)
	}

	pull, err := client.PullRequest(myRepo, args[0], nil)
	if err != nil {
		exitError(err)
	}

	comments, err := client.Comments(myRepo, args[0], nil)
	if err != nil {
		exitError(err)
	}

	f, err := ioutil.TempFile("", "barbara-edit")
	if err != nil {
		exitError(err)
	}

	color.Output = f

	line()
	color.New(color.FgBlue).Printf("From: %s\n", pull.User.Login)
	color.New(color.FgBlue).Printf("Title: %s\n", pull.Title)
	color.New(color.FgBlue).Printf("Number: %d\n", pull.Number)
	color.New(color.FgBlue).Printf("State: %s\n", pull.State)
	color.New(color.FgBlue).Printf("URL: %s\n", pull.URL)
	line()
	fmt.Fprintln(f, pull.Body)

	for _, comment := range comments {
		fmt.Fprintln(f)
		line()
		color.New(color.FgWhite).Printf("From: %s\n", comment.User.Login)
		color.New(color.FgWhite).Printf("Date: %s\n", comment.CreatedAt.Local())
		line()
		fmt.Fprintln(f)
		fmt.Fprintln(f, comment.Body)
	}

	fmt.Fprintln(f)

	f.Close()
	defer os.Remove(f.Name())

	if err := runProgram("less", "-R", f.Name()); err != nil {
		exitError(err)
	}
}

func listPRs(ctx *cli.Context) {
	client := getClient()
	client.WithToken(os.Getenv("GITHUB_TOKEN"))

	myRepo, err := repo()
	if err != nil {
		exitError(err)
	}

	pulls, err := getPRs(client, myRepo, ctx.String("state"), ctx.String("direction"), ctx.String("sort-by"))
	if err != nil {
		exitError(err)
	}

	f, err := ioutil.TempFile("", "barbara-edit")
	if err != nil {
		exitError(err)
	}

	color.Output = f

	for _, pull := range pulls {
		color.New(color.FgWhite).Printf("[ %d ] ", pull.Number)
		color.New(color.FgBlue).Printf("(%s) ", pull.User.Login)
		fmt.Fprintf(f, "%s\n", pull.Title)
	}

	f.Close()
	defer os.Remove(f.Name())

	if err := runProgram("less", "-R", f.Name()); err != nil {
		exitError(err)
	}
}
