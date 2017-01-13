package main

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"

	"github.com/crosbymichael/octokat"
	"github.com/docker/docker/pkg/term"
	"github.com/fatih/color"
	"github.com/kr/pty"
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

func replyPR(ctx *cli.Context) {
	client := getClient()
	if len(ctx.Args()) != 2 {
		exitError(errors.New("invalid arguments"))
	}

	args := ctx.Args()
	myRepo := repo(args[0])

	f, err := ioutil.TempFile("", "barbara-edit")
	if err != nil {
		exitError(err)
	}
	f.Close()
	defer os.Remove(f.Name())

	size, err := term.GetWinsize(0)
	if err != nil {
		exitError(err)
	}

	state, err := term.SetRawTerminal(0)
	if err != nil {
		exitError(err)
	}
	defer term.RestoreTerminal(0, state)

	cmd := exec.Command(os.Getenv("EDITOR"), f.Name())
	tty, err := pty.Start(cmd)
	if err != nil {
		exitError(err)
	}

	if err := term.SetWinsize(tty.Fd(), size); err != nil {
		exitError(err)
	}

	go io.Copy(tty, os.Stdin)
	go io.Copy(os.Stdout, tty)

	if err := cmd.Wait(); err != nil {
		exitError(err)
	}

	tty.Close()

	if err := term.RestoreTerminal(0, state); err != nil {
		exitError(err)
	}

	content, err := ioutil.ReadFile(f.Name())
	if err != nil {
		exitError(err)
	}

	_, err = client.AddComment(myRepo, args[1], string(content))
	if err != nil {
		exitError(err)
	}

	fmt.Printf("Comment on ticket %s posted!\n", args[1])
}

func singlePR(ctx *cli.Context) {
	client := getClient()

	if len(ctx.Args()) != 2 {
		exitError(errors.New("invalid arguments"))
	}

	args := ctx.Args()

	myRepo := repo(args[0])

	pull, err := client.PullRequest(myRepo, args[1], nil)
	if err != nil {
		exitError(err)
	}

	comments, err := client.Comments(myRepo, args[1], nil)
	if err != nil {
		exitError(err)
	}

	color.New(color.FgBlue).Printf("From: %s\n", pull.User.Login)
	color.New(color.FgBlue).Printf("Title: %s\n", pull.Title)
	color.New(color.FgBlue).Printf("Number: %d\n", pull.Number)
	color.New(color.FgBlue).Printf("State: %s\n", pull.State)
	color.New(color.FgBlue).Printf("URL: %s\n", pull.URL)
	line()
	fmt.Println(pull.Body)

	for _, comment := range comments {
		fmt.Println()
		line()
		color.New(color.FgWhite).Printf("From: %s\n", comment.User.Login)
		color.New(color.FgWhite).Printf("Date: %s\n", comment.CreatedAt.Local())
		line()
		fmt.Println()
		fmt.Println(comment.Body)
	}
}

func listPRs(ctx *cli.Context) {
	client := getClient()
	client.WithToken(os.Getenv("GITHUB_TOKEN"))

	if len(ctx.Args()) != 1 {
		exitError(errors.New("invalid arguments"))
	}

	args := ctx.Args()
	myRepo := repo(args[0])

	pulls, err := getPRs(client, myRepo, ctx.String("state"), ctx.String("direction"), ctx.String("sort-by"))
	if err != nil {
		exitError(err)
	}

	for _, pull := range pulls {
		color.New(color.FgWhite).Printf("[ %d ] ", pull.Number)
		color.New(color.FgBlue).Printf("(%s) ", pull.User.Login)
		fmt.Printf("%s\n", pull.Title)
	}
}
