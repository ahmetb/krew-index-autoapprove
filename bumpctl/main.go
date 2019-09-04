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

	err = bump.IsBumpPatch(b)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Fprintf(os.Stderr, "supplied patch is a straightforward bump")
}
