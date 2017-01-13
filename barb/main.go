package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/crosbymichael/octokat"
	"github.com/docker/docker/pkg/term"
	"github.com/fatih/color"
	"github.com/gizak/termui"
	"github.com/urfave/cli"
)

func main() {
	app := cli.NewApp()
	app.Commands = []cli.Command{
		{
			Name:        "ui",
			Description: "Launch the github UI",
			Usage:       "Launch the github UI",
			Action:      launchUI,
		},
		{
			Name:        "prs",
			Description: "Manipulate PRs",
			Usage:       "Manipulate PRs",
			Subcommands: []cli.Command{
				{
					Name:        "list",
					Description: "List PRs",
					Usage:       "List PRs",
					Action:      listPRs,
				},
				{
					Name:        "get",
					Description: "Get a PR",
					Action:      singlePR,
				},
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		exitError(err)
	}
}

func exitError(err error) {
	fmt.Fprint(os.Stderr, err.Error())
	os.Exit(1)
}

func getPRs(client *octokat.Client, state string) ([]*octokat.PullRequest, error) {
	prs, err := client.PullRequests(octokat.Repo{UserName: "docker", Name: "docker"}, nil)
	if err != nil {
		return nil, err
	}

	newPulls := []*octokat.PullRequest{}

	for _, pull := range prs {
		if pull.State == state {
			newPulls = append(newPulls, pull)
		}
	}

	return newPulls, nil
}

func line() {
	size, err := term.GetWinsize(0)
	if err != nil {
		exitError(err)
	}

	color.New(color.FgYellow, color.Bold).Println(strings.Repeat("-", int(size.Width)))
}

func singlePR(ctx *cli.Context) {
	client := octokat.NewClient()

	pull, err := client.PullRequest(octokat.Repo{UserName: "docker", Name: "docker"}, ctx.Args()[0], nil)
	if err != nil {
		exitError(err)
	}

	comments, err := client.Comments(octokat.Repo{UserName: "docker", Name: "docker"}, ctx.Args()[0], nil)
	if err != nil {
		exitError(err)
	}

	fmt.Printf("From: %s\n", pull.User.Login)
	fmt.Printf("Title: %s\n", pull.Title)
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
	client := octokat.NewClient()
	client.WithToken(os.Getenv("GITHUB_TOKEN"))

	pulls, err := getPRs(client, "open")
	if err != nil {
		exitError(err)
	}

	for _, pull := range pulls {
		color.New(color.FgWhite).Printf("[ %d ] ", pull.Number)
		color.New(color.FgBlue).Printf("(%s) ", pull.User.Login)
		fmt.Printf("%s\n", pull.Title)
	}
}

func launchUI(ctx *cli.Context) {
	list := termui.NewList()
	client := octokat.NewClient()

	fmt.Println("Loading pull requests, please wait...")
	pulls, err := getPRs(client, "open")
	if err != nil {
		exitError(err)
	}

	titles := []string{}
	for _, pull := range pulls {
		titles = append(titles, fmt.Sprintf("[ %d ] %q", pull.Number, pull.Title))
	}

	list.Border = true
	list.Items = titles
	list.ItemFgColor = termui.ColorWhite

	if err := termui.Init(); err != nil {
		exitError(err)
	}
	defer termui.Close()

	list.Height = termui.TermHeight()
	list.Width = termui.TermWidth()
	termui.Body.AddRows(termui.NewRow(termui.NewCol(0, 12, list)))

	termui.Handle("/sys/kbd/q", func(termui.Event) {
		termui.StopLoop()
	})

	termui.Render(termui.Body)
	termui.Loop()
}
