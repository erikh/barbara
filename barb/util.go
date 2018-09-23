package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"regexp"
	"strings"

	"github.com/docker/docker/pkg/term"
	"github.com/fatih/color"
	"github.com/google/go-github/github"
	"github.com/kr/pty"
	"golang.org/x/oauth2"
)

var urlRegexp = regexp.MustCompile(`(https://|git@)github.com[:/](\S+)`)

func exitError(err error) {
	fmt.Fprintln(os.Stderr, err.Error())
	os.Exit(1)
}

func repo() (string, string, error) {
	content, err := exec.Command("git", "config", "--get", "remote.origin.url").CombinedOutput()
	if err != nil {
		return "", "", err
	}

	match := urlRegexp.FindStringSubmatch(string(content))
	if len(match) != 3 {
		return "", "", errors.New("invalid url in origin remote")
	}

	parts := strings.Split(match[2], "/")
	return parts[0], parts[1], nil
}

func line() {
	size, err := term.GetWinsize(0)
	if err != nil {
		exitError(err)
	}

	color.New(color.FgYellow, color.Bold).Println(strings.Repeat("-", int(size.Width)))
}

func getClient() *github.Client {
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: os.Getenv("GITHUB_TOKEN")},
	)
	tc := oauth2.NewClient(ctx, ts)

	return github.NewClient(tc)
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
