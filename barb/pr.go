package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/crosbymichael/octokat"
	"github.com/fatih/color"
	"github.com/urfave/cli"
)

func getPRs(client *octokat.Client, ctx *cli.Context, repo octokat.Repo) ([]*octokat.PullRequest, error) {
	newPulls := []*octokat.PullRequest{}

	for page := 1; page < ctx.Int("max-pages"); page++ {
		params := map[string]string{
			"state":     ctx.String("state"),
			"direction": ctx.String("direction"),
			"sort":      ctx.String("sort-by"),
			"page":      fmt.Sprintf("%d", page),
		}

		prs, err := client.PullRequests(repo, &octokat.Options{QueryParams: params})
		if err != nil {
			return nil, err
		}

		if len(prs) == 0 {
			break
		}

		for _, pull := range prs {
			newPulls = append(newPulls, pull)
		}
	}

	return newPulls, nil
}

func diffPR(ctx *cli.Context) {
	client := getClient()

	args := ctx.Args()
	if len(args) != 1 {
		exitError(errors.New("invalid arguments"))
	}

	myRepo, err := repo()
	if err != nil {
		exitError(err)
	}

	prfs, err := client.PullRequestFiles(myRepo, args[0], nil)
	if err != nil {
		exitError(err)
	}

	f, err := ioutil.TempFile("", "barbara-edit")
	if err != nil {
		exitError(err)
	}
	defer os.Remove(f.Name())

	color.Output = f

	for _, file := range prfs {
		line()
		fmt.Fprintln(f, file.FileName)
		line()

		for _, line := range strings.Split(file.Patch, "\n") {
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
	if err := runProgram("less", "-R", f.Name()); err != nil {
		exitError(err)
	}
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
	defer os.Remove(f.Name())

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

func listPRs(ctx *cli.Context) {
	client := getClient()
	client.WithToken(os.Getenv("GITHUB_TOKEN"))

	myRepo, err := repo()
	if err != nil {
		exitError(err)
	}

	pulls, err := getPRs(client, ctx, myRepo)
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
