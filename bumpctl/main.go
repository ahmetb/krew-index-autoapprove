package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"krew-index-autoapprove/bump"
	"log"
	"os"
)

var (
	file string
)

func init() {
	flag.StringVar(&file, "file", "", "pull request patch")
	flag.Parse()
}

func main() {
	if file == "" {
		log.Fatal("-file is not set")
	}

	b, err := ioutil.ReadFile(file)
	if err != nil {
		log.Fatal(err)
	}

	ok, err := bump.IsBumpPatch(b)
	if err != nil {
		log.Fatal(err)
	}
	if !ok {
		log.Fatal("patch doesn't seem to be a bump PR")
	}

	err = bump.IsValidBump(b)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Fprintf(os.Stderr, "supplied patch is a straightforward bump")
}
