package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/google/go-github/v28/github"
	"golang.org/x/oauth2"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"krew-index-autoapprove/bump"
)

var (
	gh *github.Client
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		log.Fatal("$PORT env var is not set")
	}

	token := os.Getenv("GITHUB_TOKEN")
	if token == "" {
		log.Fatal("GITHUB_TOKEN is not set")
	}
	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
	gh = github.NewClient(oauth2.NewClient(context.TODO(), ts))

	http.HandleFunc("/webhook", webhook)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func webhook(w http.ResponseWriter, req *http.Request) {
	if req.Method != "POST" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	evType := req.Header.Get("X-GitHub-Event")
	if evType != "pull_request" {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "expected X-GitHub-Event: pull_request; got: %s", evType)
		return
	}

	type prEvent struct {
		Action      string `json:"action"`
		PullRequest struct {
			URL     string `json:"url"`
			Number  int    `json:"number"`
			DiffURL string `json:"diff_url"`
		} `json:"pull_request"`
		Repository struct {
			Name  string `json:"name"`
			Owner struct {
				Login string `json:"login"`
			} `json:"owner"`
		} `json:"repository"`
	}

	if req.Body != nil {
		defer req.Body.Close()
	}
	var ev prEvent
	if err := json.NewDecoder(req.Body).Decode(&ev); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "error parsing json payload in body: %v", err)
		return
	}

	if ev.Action != "opened" && ev.Action != "create" && ev.Action != "synchronize" {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "event action %q not accepted", ev.Action)
		return
	}

	patchReq, err := http.Get(ev.PullRequest.DiffURL)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "failed to get patch: %v", err)
		return
	}
	defer patchReq.Body.Close()
	b, err := ioutil.ReadAll(patchReq.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "failed to read patch url: %v", err)
		return
	}

	ok, err := bump.IsBumpPatch(b)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "error determining if patch is a bump: %v", err)
		return
	}

	if !ok {
		w.WriteHeader(http.StatusPreconditionFailed)
		fmt.Fprintf(w, "patch is not a bump pr")
		return
	}

	err = bump.IsValidBump(b)

	var comment string
	comment += ":robot: _Beep beep! I’m a robot speaking on behalf of @ahmetb._ :robot:\n\n-----\n\n"
	if err == nil {
		comment += "This pull request seems to be a straightforward version bump.\n"
		comment += "I'll go ahead and accept it. :+1: Cheers.\n\n"
		comment += "/lgtm\n"
		comment += "/approve\n"
	} else {
		comment += "This pull request **doesn't** seem to be a straightforward version bump." +
			" I'll have a human review this.\n\n"
		comment += "_Why wasn't this detected as a plugin version bump:_\n\n>" + err.Error()
	}

	_, resp, err := gh.Issues.CreateComment(context.TODO(),
		ev.Repository.Owner.Login,
		ev.Repository.Name,
		ev.PullRequest.Number,
		&github.IssueComment{
			Body: &comment,
		})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "github commenting error: %v\n", err)
		fmt.Fprintf(w, "resp status from github API: %d\n", resp.StatusCode)
		fmt.Fprintf(w, "resp headers github API: %v\n", resp.Header)
		return
	}
	w.WriteHeader(http.StatusAccepted)
	return
}
