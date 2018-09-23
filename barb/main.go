package main

import (
	"os"

	"github.com/urfave/cli"
)

func main() {
	app := cli.NewApp()
	app.Usage = "barbara is a github client"
	app.Version = "0.1.0"
	app.Commands = []cli.Command{
		{
			Name:      "issue",
			ShortName: "i",
			Usage:     "Subcommand trampoline for all issues",
			Subcommands: []cli.Command{
				{
					Name:      "get",
					Usage:     "get info on a single issue",
					ArgsUsage: "[id]",
					Action:    getIssue,
				},
				{
					Name:      "reply",
					Usage:     "Reply to an issue. Spawns $EDITOR",
					ArgsUsage: "[id]",
					Action:    replyIssue,
				},
				{
					Name:      "list",
					Usage:     "list issues",
					ArgsUsage: "",
					Action:    listIssue,
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
			},
		},
		{
			Name:  "pr",
			Usage: "Subcommand trampoline for pull requests",
			Subcommands: []cli.Command{
				{
					Name:      "close",
					ShortName: "c",
					Usage:     "Close a PR",
					ArgsUsage: "[pull request id]",
					Action:    closePR,
				},
				{
					Name:      "watch-hooks",
					ShortName: "w",
					Usage:     "Watch for hooks (like CI) to complete for a PR",
					Action:    watch,
				},
				{
					Name:   "get",
					Usage:  "Get state/comments for an PR",
					Action: get,
				},
				{
					Name:   "reply",
					Usage:  "Reply to a ticket",
					Action: reply,
				},
				{
					Name:   "list",
					Usage:  "List PRs",
					Action: listPRs,
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
					Name:  "create",
					Usage: "Create a PR",
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
					Name:   "merge",
					Usage:  "Merge a PR",
					Action: mergePR,
				},
				{
					Name:   "diff",
					Usage:  "Get the diff for a PR",
					Action: diffPR,
				},
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		exitError(err)
	}
}
