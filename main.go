package main

import (
	"fmt"
	"os"

	"github.com/go-git/go-git/v5"
)

func main() {
	path, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	repo, err := git.PlainOpen(path)
	if err != nil {
		panic(err)
	}

	branches, err := localBranches()
	if err != nil {
		panic(err)
	}

	fmt.Printf("%v\n", repo)
	fmt.Printf("%v\n", branches)
}
