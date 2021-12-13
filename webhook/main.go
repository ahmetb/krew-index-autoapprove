package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/google/go-github/v28/github"
	"golang.org/x/oauth2"

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
			User    struct {
				Login string `json:"login"`
			} `json:"user"`
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

	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "cannot read request body: %v", err)
	}

	var ev prEvent
	if err := json.Unmarshal(body, &ev); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "error parsing json payload in body: %v\n", err)
		fmt.Fprintf(w, "payload: %s", string(body))
		return
	}

	if ev.Action != "opened" {
		// action="opened" occurs when a PR is first created
		// action="synchronize" occurs when a new commit is pushed
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "event action %q not accepted", ev.Action)
		return
	}

	patchReq, err := http.Get(ev.PullRequest.DiffURL)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "failed to get patch: %v\n", err)
		fmt.Fprintf(w, "patch URL: %s\n", ev.PullRequest.DiffURL)
		fmt.Fprintf(w, "payload: %s", string(body))
		return
	}
	defer patchReq.Body.Close()
	patch, err := ioutil.ReadAll(patchReq.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "failed to read patch url: %v\n", err)
		fmt.Fprintf(w, "payload: %s", string(body))
		return
	}

	isPluginBump, err := bump.IsBumpPatch(patch)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "error determining if patch is a bump: %v\n", err)
		fmt.Fprintf(w, "patch: %s", string(patch))
		return
	}

	isNewSubmission, err := bump.IsNewPluginSubmission(patch)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "error determining if patch is a new plugin submission: %v\n", err)
		fmt.Fprintf(w, "patch: %s", string(patch))
		return
	}

	if !isPluginBump && !isNewSubmission {
		w.WriteHeader(http.StatusPreconditionFailed)
		fmt.Fprintf(w, "patch is not a new plugin or a version bump pr\n")
		fmt.Fprintf(w, "patch: %s", string(patch))
		return
	}

	var comment string
	comment += ":robot: _Beep beep! I’m a robot speaking on behalf of @ahmetb._ :robot:\n\n-----\n\n"

	if isPluginBump {
		err = bump.IsValidBump(patch)
		if err == nil {
			comment += "This pull request seems to be a straightforward version bump. "
			comment += "I'll go ahead and accept it. :+1: Cheers.\n\n"
			comment += "/lgtm\n"
			comment += "/approve\n"
		} else {
			comment += "This pull request **does not** seem to be a straightforward version bump." +
				" I'll have a human review this. If we don't respond within a day, feel free to ping us.\n\n"
			comment += "_Why wasn't this detected as a plugin version bump:_\n\n>" + err.Error()
		}
		comment += "\n/kind plugin-update"
	} else if isNewSubmission {
		if err := bump.IsReviewablePluginSubmission(patch); err != nil {
			comment += "This pull request will not be reviewed. Reason:\n\n"
			comment += "> " + err.Error() + "\n\n"
			comment += "/close"
		} else {
			if ev.PullRequest.User.Login == "krew-release-bot" {
				comment += "First plugin submission should not be done through the @krew-release-bot "
				comment += "as it is not possible to iterate on this pull request based on feedback.\n\n"
				comment += "Please submit a new pull request yourself.\n"
				comment += "/close"
			} else {
				comment += "Thanks for submitting your kubectl plugin to Krew!\n"
				comment += "One of the krew-index maintainers will review it soon. "
				comment += "Note that the reviews for new plugin submissions may take a few days.\n\n"
				comment += "In the meanwhile, here are a few tips to make your plugin manifest better:\n"
				comment += "* Make sure your plugin follows the [best practices](https://krew.sigs.k8s.io/docs/developer-guide/develop/best-practices/).\n"
				comment += "* Eliminate redundant wording form `shortDescription` (it should be max 50 characters).\n"
				comment += "* Try to word wrap your `description` to 80-character lines (no usage examples, please).\n"
				comment += "\nThanks for your patience!"
			}
		}
		comment += "\n/kind new-plugin"
	} else {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintln(w, "unhandled situation")
		return
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
