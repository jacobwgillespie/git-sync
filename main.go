package main

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/jacobwgillespie/git-sync/git"
)

var (
	green      = "\033[32m"
	lightGreen = "\033[32;1m"
	red        = "\033[31m"
	lightRed   = "\033[31;1m"
	resetColor = "\033[0m"
)

func main() {
	remote, err := git.MainRemote()
	check(err)

	defaultBranch := git.BranchShortName(git.DefaultBranch(remote))
	fullDefaultBranch := fmt.Sprintf("refs/remotes/%s/%s", remote, defaultBranch)
	currentBranch := ""
	if current, err := git.CurrentBranch(); err == nil {
		currentBranch = git.BranchShortName(current)
	}

	err = git.Spawn("fetch", "--prune", "--quiet", "--progress", remote)
	check(err)

	branchToRemote := map[string]string{}
	if lines, err := git.ConfigAll("branch.*.remote"); err == nil {
		configRe := regexp.MustCompile(`^branch\.(.+?)\.remote (.+)`)

		for _, line := range lines {
			if matches := configRe.FindStringSubmatch(line); len(matches) > 0 {
				branchToRemote[matches[1]] = matches[2]
			}
		}
	}

	branches, err := git.LocalBranches()
	check(err)

	for _, branch := range branches {
		fullBranch := fmt.Sprintf("refs/heads/%s", branch)
		remoteBranch := fmt.Sprintf("refs/remotes/%s/%s", remote, branch)
		gone := false

		if branchToRemote[branch] == remote {
			if upstream, err := git.SymbolicFullName(fmt.Sprintf("%s@{upstream}", branch)); err == nil {
				remoteBranch = upstream
			} else {
				remoteBranch = ""
				gone = true
			}
		} else if !git.HasFile(strings.Split(remoteBranch, "/")...) {
			remoteBranch = ""
		}

		if remoteBranch != "" {
			diff, err := git.NewRange(fullBranch, remoteBranch)
			check(err)

			if diff.IsIdentical() {
				continue
			} else if diff.IsAncestor() {
				if branch == currentBranch {
					git.Quiet("merge", "--ff-only", "--quiet", remoteBranch)
				} else {
					git.Quiet("update-ref", fullBranch, remoteBranch)
				}
				fmt.Printf("%sUpdated branch %s%s%s (was %s).\n", green, lightGreen, branch, resetColor, diff.A[0:7])
			} else {
				fmt.Fprintf(os.Stderr, "warning: '%s' seems to contain unpushed commits\n", branch)
			}
		} else if gone {
			diff, err := git.NewRange(fullBranch, fullDefaultBranch)
			check(err)

			if diff.IsAncestor() {
				if branch == currentBranch {
					git.Quiet("checkout", "--quiet", defaultBranch)
					currentBranch = defaultBranch
				}
				git.Quiet("branch", "-D", branch)
				fmt.Printf("%sDeleted branch %s%s%s (was %s).\n", red, lightRed, branch, resetColor, diff.A[0:7])
			} else {
				fmt.Fprintf(os.Stderr, "warning: '%s' was deleted on %s, but appears not merged into '%s'\n", branch, remote, defaultBranch)
			}
		}
	}
}

func check(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
}
