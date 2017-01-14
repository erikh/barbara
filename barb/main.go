package main

import (
	"os"

	"github.com/urfave/cli"
)

func main() {
	app := cli.NewApp()
	app.Commands = []cli.Command{
		{
			Name:        "watch-hooks",
			Description: "Watch for hooks (like CI) to complete for a PR",
			Usage:       "Watch for hooks (like CI) to complete for a PR",
			Action:      watch,
		},
		{
			Name:        "get",
			Description: "Get state/comments for an PR",
			Usage:       "Get state/comments for an PR",
			Action:      get,
		},
		{
			Name:        "reply",
			Description: "Reply to a ticket",
			Usage:       "Reply to a ticket",
			Action:      reply,
		},
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
				cli.IntFlag{
					Name:  "m, max-pages",
					Usage: "Maximum number of list pages to fetch",
					Value: 5,
				},
			},
		},
		{
			Name:        "create",
			Description: "Create a PR",
			Usage:       "Create a PR",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name: "t, title",
				},
				cli.StringFlag{
					Name:  "b, base",
					Value: "master",
				},
			},
			Action: createPR,
		},
		{
			Name:        "merge",
			Description: "Merge a PR",
			Usage:       "Merge a PR",
			Action:      mergePR,
		},
		{
			Name:        "diff",
			Description: "Get the diff for a PR",
			Usage:       "Get the diff for a PR",
			Action:      diffPR,
		},
	}

	if err := app.Run(os.Args); err != nil {
		exitError(err)
	}
}
