package main

import (
	"fmt"
	"gopkg.in/go-playground/webhooks.v5/github"
	"net/http"
	"os"
)

const (
	path = "/hooks"
)

var (
	hookSecret = os.Getenv("HOOK_SECRET")
)

func main() {

	http.HandleFunc(path, webhookHandler)
	http.ListenAndServe(":4444", nil)
}

func webhookHandler(w http.ResponseWriter, r *http.Request) {
	hook, _ := github.New(github.Options.Secret(hookSecret))
	payload, err := hook.Parse(r, github.PullRequestEvent, github.PushEvent)
	if err != nil {
		if err == github.ErrEventNotFound {
			fmt.Println("does this work?")
		}
	}
	switch payload.(type) {
	case github.PullRequestPayload:
		pullRequest := payload.(github.PullRequestPayload)
		// Do whatever you want from here...
		fmt.Printf("%+v", pullRequest)
		
	case github.PushPayload:
		push := payload.(github.PushPayload)
		fmt.Printf("%+v", push)
	}
}
