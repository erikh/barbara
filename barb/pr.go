package main

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/google/go-github/github"
	"github.com/urfave/cli"
)

func getPRs(client *github.Client, ctx *cli.Context, owner, repo string) ([]*github.PullRequest, error) {
	newPulls := []*github.PullRequest{}

	for page := 1; page < ctx.Int("max-pages"); page++ {
		params := &github.PullRequestListOptions{
			State:     ctx.String("state"),
			Sort:      ctx.String("sort-by"),
			Direction: ctx.String("direction"),
			ListOptions: github.ListOptions{
				Page: page,
			},
		}

		prs, _, err := client.PullRequests.List(context.Background(), owner, repo, params)
		if err != nil {
			return nil, err
		}

		if len(prs) == 0 {
			break
		}

		newPulls = append(newPulls, prs...)
	}

	return newPulls, nil
}

func diffPR(ctx *cli.Context) {
	client := getClient()

	args := ctx.Args()
	if len(args) != 1 {
		exitError(errors.New("invalid arguments"))
	}

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

	commits, _, err := client.Repositories.CompareCommits(context.Background(), owner, repo, pr.Base.GetSHA(), pr.Head.GetSHA())

	f, err := ioutil.TempFile("", "barbara-edit")
	if err != nil {
		exitError(err)
	}
	defer os.Remove(f.Name())

	color.Output = f

	for _, file := range commits.Files {
		line()
		fmt.Fprintln(f, file.GetFilename())
		line()

		for _, line := range strings.Split(file.GetPatch(), "\n") {
			switch line[0] {
			case '+':
				color.New(color.FgGreen).Println(line)
			case '-':
				color.New(color.FgRed).Println(line)
			case '!':
				color.New(color.FgYellow).Println(line)
			default:
				fmt.Fprintln(f, line)
			}
		}
	}

	f.Close()
}

func closePR(ctx *cli.Context) {
	client := getClient()
	args := ctx.Args()

	if len(args) != 1 {
		exitError(errors.New("invalid arguments"))
	}

	owner, repo, err := repo()
	if err != nil {
		exitError(err)
	}

	num, err := strconv.Atoi(args[0])
	if err != nil {
		exitError(err)
	}

	now := time.Now()
	_, _, err = client.PullRequests.Edit(context.Background(), owner, repo, num, &github.PullRequest{State: github.String("closed"), ClosedAt: &now})
	if err != nil {
		exitError(err)
	}

	fmt.Printf("Pull request %s closed!\n", args[0])
}

func mergePR(ctx *cli.Context) {
	client := getClient()
	args := ctx.Args()
	if len(args) != 1 {
		exitError(errors.New("invalid arguments"))
	}

	owner, repo, err := repo()
	if err != nil {
		exitError(err)
	}

	num, err := strconv.Atoi(args[0])
	if err != nil {
		exitError(err)
	}

	_, _, err = client.PullRequests.Merge(context.Background(), owner, repo, num, "", nil)
	if err != nil {
		exitError(err)
	}

	fmt.Printf("PR #%s successfully merged!\n", args[0])
}

func createPR(ctx *cli.Context) {
	client := getClient()

	args := ctx.Args()

	git, err := exec.LookPath("git")
	if err != nil {
		exitError(err)
	}

	out, err := exec.Command(git, "log", "-n", "1").Output()
	if err != nil {
		exitError(err)
	}

	if len(args) != 1 {
		exitError(errors.New("invalid arguments"))
	}

	lines := strings.Split(string(out), "\n")
	trimmed := []string{}

	for _, l := range lines {
		trimmed = append(trimmed, strings.TrimSpace(l))
	}

	title := ctx.String("title")
	if title == "" {
		title = trimmed[4]
	}

	owner, repo, err := repo()
	if err != nil {
		exitError(err)
	}

	f, err := ioutil.TempFile("", "barbara-edit")
	if err != nil {
		exitError(err)
	}
	f.Write([]byte(strings.Join(trimmed[6:], "\n")))
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
		exitError(errors.New("prs must have content"))
	}

	pr, _, err := client.PullRequests.Create(context.Background(), owner, repo, &github.NewPullRequest{
		Title: github.String(title),
		Body:  github.String(string(content)),
		Base:  github.String(ctx.String("base")),
		Head:  github.String(args[0]),
	})

	if err != nil {
		exitError(err)
	}

	fmt.Printf("PR %d created!\n", pr.GetNumber())
}

func listPRs(ctx *cli.Context) {
	client := getClient()

	owner, repo, err := repo()
	if err != nil {
		exitError(err)
	}

	pulls, err := getPRs(client, ctx, owner, repo)
	if err != nil {
		exitError(err)
	}

	color.Output = os.Stdout

	for _, pull := range pulls {
		color.New(color.FgWhite).Printf("[ %d ] ", pull.GetNumber())
		color.New(color.FgBlue).Printf("(%s) ", pull.User.GetLogin())
		fmt.Fprintf(os.Stdout, "%s", pull.GetTitle())

		status, _, err := client.Repositories.GetCombinedStatus(context.Background(), owner, repo, pull.Head.GetSHA(), nil)
		if err != nil {
			exitError(err)
		}

		var stateColor *color.Color

		switch status.GetState() {
		case "success":
			stateColor = color.New(color.FgGreen)
		case "pending":
			stateColor = color.New(color.FgWhite)
		case "error":
			stateColor = color.New(color.FgYellow)
		case "failure":
			stateColor = color.New(color.FgRed)
		}

		stateColor.Printf(" [ %s ]", status.GetState())
		color.New(color.Reset).Print("\n")
	}
}
