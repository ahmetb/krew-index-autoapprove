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
	file   string
	author string
)

func init() {
	flag.StringVar(&file, "file", "", "pull request patch")
	flag.StringVar(&author, "author", "", "pull request author (for plugin list updates)")
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

	if author != "" {
		ok, err := bump.IsPluginListUpdate(author, b)
		if err != nil {
			log.Fatal(err)
		}
		if ok {
			fmt.Fprintf(os.Stderr, "supplied patch is a straightforward plugin list update")
			return
		}
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
