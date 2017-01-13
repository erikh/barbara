package main

import (
	"fmt"
	"os"

	"github.com/urfave/cli"
)

func main() {
	app := cli.NewApp()
	app.Commands = []cli.Command{
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
					Flags: []cli.Flag{
						cli.StringFlag{
							Name:  "s, state",
							Usage: "State of prs (open, closed)",
							Value: "open",
						},
						cli.StringFlag{
							Name:  "b, sort-by",
							Usage: "Sort by this value",
							Value: "created",
						},
						cli.StringFlag{
							Name:  "d, direction",
							Usage: "Direction of sort",
							Value: "desc",
						},
					},
				},
				{
					Name:        "get",
					Description: "Get a PR",
					Usage:       "Get a PR",
					Action:      singlePR,
				},
				{
					Name:        "reply",
					Description: "Reply to a a ticket",
					Usage:       "Reply to a a ticket",
					Action:      replyPR,
				},
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		exitError(err)
	}
}

func exitError(err error) {
	fmt.Fprintln(os.Stderr, err.Error())
	os.Exit(1)
}
