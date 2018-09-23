package main

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"time"

	"github.com/fatih/color"
	"github.com/google/go-github/github"
	"github.com/urfave/cli"
)

func watch(ctx *cli.Context) {
	client := getClient()

	args := ctx.Args()
	if len(args) < 1 {
		exitError(errors.New("invalid arguments"))
	}

	owner, repo, err := repo()
	if err != nil {
		exitError(err)
	}

	color.New(color.FgHiWhite).Printf("Monitoring PRs for %s/%s (ids: %v); will output as things finish.\n", owner, repo, args)

	doneChan := make(chan []string, len(args))

	for _, arg := range args {
		go func(arg string) {
			for {
				num, err := strconv.Atoi(arg)
				if err != nil {
					exitError(err)
				}

				pr, _, err := client.PullRequests.Get(context.Background(), owner, repo, num)
				if err != nil {
					exitError(err)
				}

				status, _, err := client.Repositories.GetCombinedStatus(context.Background(), owner, repo, pr.Head.GetSHA(), nil)
				if err != nil {
					exitError(err)
				}

				if status.GetState() != "pending" {
					doneChan <- []string{arg, status.GetState()}
				}

				time.Sleep(30 * time.Second)
			}
		}(arg)
	}

	var i int

	for params := range doneChan {
		i++
		fmt.Printf("Finished: %v (%v)\n", params[0], params[1])
		fmt.Printf("Remaining: %d\n", i-len(args))

		if i == len(args) {
			return
		}
	}
}

func reply(ctx *cli.Context) {
	client := getClient()
	if len(ctx.Args()) != 1 {
		exitError(errors.New("invalid arguments"))
	}

	args := ctx.Args()
	owner, repo, err := repo()
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

	if len(content) == 0 {
		exitError(errors.New("no content to post"))
	}

	num, err := strconv.Atoi(args[0])
	if err != nil {
		exitError(err)
	}

	_, _, err = client.PullRequests.CreateComment(context.Background(), owner, repo, num, &github.PullRequestComment{Body: github.String(string(content))})
	if err != nil {
		exitError(err)
	}

	fmt.Printf("Comment on ticket %s posted!\n", args[0])
}

func get(ctx *cli.Context) {
	client := getClient()

	if len(ctx.Args()) != 1 {
		exitError(errors.New("invalid arguments"))
	}

	args := ctx.Args()

	owner, repo, err := repo()
	if err != nil {
		exitError(err)
	}

	num, err := strconv.Atoi(args[0])
	if err != nil {
		exitError(err)
	}

	pr, _, err := client.PullRequests.Get(context.Background(), owner, repo, num)
	if err != nil {
		exitError(err)
	}

	allComments := []*github.PullRequestComment{}
	comments := []*github.PullRequestComment{}

	for i := 0; len(comments) != 0; i++ {
		comments, _, err = client.PullRequests.ListComments(context.Background(), owner, repo, num, &github.PullRequestListCommentsOptions{
			ListOptions: github.ListOptions{
				Page:    i,
				PerPage: 100,
			},
		})
		if err != nil {
			exitError(err)
		}

		allComments = append(allComments, comments...)
	}

	color.Output = os.Stdout

	line()
	color.New(color.FgHiBlue).Printf("From: %s\n", pr.User.GetLogin())
	color.New(color.FgHiBlue).Printf("Last Updated: %v\n", pr.GetUpdatedAt())
	color.New(color.FgHiBlue).Printf("Title: %s\n", pr.GetTitle())
	color.New(color.FgHiBlue).Printf("Number: %d\n", pr.GetNumber())
	color.New(color.FgHiBlue).Printf("URL: %s\n", pr.GetHTMLURL())

	stateColor := color.New()
	switch pr.GetState() {
	case "open":
		stateColor = color.New(color.FgGreen)
	case "closed":
		stateColor = color.New(color.FgRed)
	}

	stateColor.Printf("State: %s\n", pr.GetState())

	status, _, err := client.Repositories.GetCombinedStatus(context.Background(), owner, repo, pr.Head.GetSHA(), nil)
	if err != nil {
		exitError(err)
	}

	switch status.GetState() {
	case "success":
		stateColor = color.New(color.FgHiGreen)
	case "pending":
		stateColor = color.New(color.FgHiWhite)
	case "error":
		stateColor = color.New(color.FgHiYellow)
	case "failure":
		stateColor = color.New(color.FgHiRed)
	}

	stateColor.Print("Hooks State: ")

	if status.GetState() == "success" {
		stateColor.Println("success")
	} else {
		stateColor.Println()

		for _, state := range status.Statuses {
			if state.GetState() != "success" {
				stateColor.Println("\t", state.GetContext(), ":", state.GetTargetURL())
			}
		}
	}

	line()
	fmt.Println(pr.GetBody())

	for _, comment := range comments {
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
