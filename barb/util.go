package main

import (
	"errors"
	"io"
	"os"
	"os/exec"
	"regexp"
	"strings"

	"github.com/crosbymichael/octokat"
	"github.com/docker/docker/pkg/term"
	"github.com/fatih/color"
	"github.com/kr/pty"
)

var urlRegexp = regexp.MustCompile(`(https://|git@)github.com[:/]([^\s]+)`)

func repo() (octokat.Repo, error) {
	repo := octokat.Repo{}
	content, err := exec.Command("git", "config", "--get", "remote.origin.url").CombinedOutput()
	if err != nil {
		return repo, err
	}

	match := urlRegexp.FindStringSubmatch(string(content))
	if len(match) != 3 {
		return repo, errors.New("invalid url in origin remote")
	}

	parts := strings.Split(match[2], "/")

	return octokat.Repo{Name: parts[1], UserName: parts[0]}, nil
}

func line() {
	size, err := term.GetWinsize(0)
	if err != nil {
		exitError(err)
	}

	color.New(color.FgYellow, color.Bold).Println(strings.Repeat("-", int(size.Width)))
}

func getClient() *octokat.Client {
	client := octokat.NewClient()
	client.WithToken(os.Getenv("GITHUB_TOKEN"))
	return client
}

func runProgram(command ...string) error {
	size, err := term.GetWinsize(0)
	if err != nil {
		return err
	}

	state, err := term.SetRawTerminal(0)
	if err != nil {
		return err
	}
	defer term.RestoreTerminal(0, state)

	cmd := exec.Command(command[0], command[1:]...)

	tty, err := pty.Start(cmd)
	if err != nil {
		return err
	}

	if err := term.SetWinsize(tty.Fd(), size); err != nil {
		return err
	}

	go io.Copy(tty, os.Stdin)
	go io.Copy(os.Stdout, tty)

	if err := cmd.Wait(); err != nil {
		return err
	}

	tty.Close()

	if err := term.RestoreTerminal(0, state); err != nil {
		return err
	}

	return nil
}
