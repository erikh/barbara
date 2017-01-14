package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"time"

	"github.com/fatih/color"
	"github.com/urfave/cli"
)

func watch(ctx *cli.Context) {
	client := getClient()

	args := ctx.Args()
	if len(args) != 1 {
		exitError(errors.New("invalid arguments"))
	}

	myRepo, err := repo()
	if err != nil {
		exitError(err)
	}

	for {
		pr, err := client.PullRequest(myRepo, args[0], nil)
		if err != nil {
			exitError(err)
		}

		status, err := client.CombinedStatus(myRepo, pr.Head.Sha, nil)
		if err != nil {
			exitError(err)
		}

		if status.State != "pending" {
			os.Exit(0)
		}

		time.Sleep(1 * time.Second)
	}
}

func reply(ctx *cli.Context) {
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

func get(ctx *cli.Context) {
	client := getClient()

	if len(ctx.Args()) != 1 {
		exitError(errors.New("invalid arguments"))
	}

	args := ctx.Args()

	myRepo, err := repo()
	if err != nil {
		exitError(err)
	}

	pr, err := client.PullRequest(myRepo, args[0], nil)
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
	color.New(color.FgBlue).Printf("From: %s\n", pr.User.Login)
	color.New(color.FgBlue).Printf("Title: %s\n", pr.Title)
	color.New(color.FgBlue).Printf("Number: %d\n", pr.Number)
	color.New(color.FgBlue).Printf("URL: %s\n", pr.HTMLURL)

	status, err := client.CombinedStatus(myRepo, pr.Head.Sha, nil)
	if err != nil {
		exitError(err)
	}

	stateColor := color.New()
	switch pr.State {
	case "open":
		stateColor = color.New(color.FgGreen)
	case "closed":
		stateColor = color.New(color.FgRed)
	}

	stateColor.Printf("State: %s\n", pr.State)

	switch status.State {
	case "success":
		stateColor = color.New(color.FgGreen)
	case "pending":
		stateColor = color.New(color.FgWhite)
	case "error":
		stateColor = color.New(color.FgYellow)
	case "failure":
		stateColor = color.New(color.FgRed)
	}

	stateColor.Print("Hooks State: ")

	if status.State == "success" {
		stateColor.Println("success")
	} else {
		stateColor.Println()

		for _, state := range status.Statuses {
			if state.State != "success" {
				stateColor.Println("\t", state.Context, ":", state.TargetURL)
			}
		}
	}

	line()
	fmt.Fprintln(f, pr.Body)

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
