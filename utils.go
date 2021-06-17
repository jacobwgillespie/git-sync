package main

import (
	"os"
	"os/exec"
	"strings"
)

func localBranches() ([]string, error) {
	output, err := execGit("branch", "--list")
	if err != nil {
		return nil, err
	}
	branches := []string{}
	for _, branch := range splitLines(output) {
		branches = append(branches, branch[2:])
	}
	return branches, nil
}

func execGit(args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	cmd.Stderr = os.Stderr
	output, err := cmd.Output()
	return string(output), err
}

func splitLines(output string) []string {
	output = strings.TrimSuffix(output, "\n")
	if output == "" {
		return []string{}
	}
	return strings.Split(output, "\n")
}
