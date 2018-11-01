package main

import (
	"bytes"
	"crypto/rsa"
	"encoding/json"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"gopkg.in/go-playground/webhooks.v5/github"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"
)

type CreateCheckRunPayload struct {
	Name        string  `json:"name"`
	HeadSha     string  `json:"head_sha"`
	DetailsURL  string  `json:"details_url"`
	ExternalID  string  `json:"external_id"`
	Status      string  `json:"status"`
	StartedAt   string  `json:"started_at"`
	Conclusion  string  `json:"conclusion"`
	CompletedAt string  `json:"completed_at"`
	Output      Output  `json:"output"`
	Actions     Actions `json:"actions"`
}

type Actions struct {
	Label       string `json:"label"`
	Description string `json:"description"`
	Identifier  string `json:"identifier"`
}

type Output struct {
	Title       string        `json:"title"`
	Summary     string        `json:"summary"`
	Text        string        `json:"text"`
	Annotations []Annotations `json:"annotations"`
	Images      []Images      `json:"images"`
}

type Images struct {
	Alt      string `json:"alt"`
	ImageURL string `json:"image_url"`
	Caption  string `json:"caption"`
}

type Annotations struct {
	Path            string `json:"path"`
	StartLine       int64  `json:"start_line"`
	EndLine         int64  `json:"end_line"`
	StartColumn     int64  `json:"start_column"`
	EndColumn       int64  `json:"end_column"`
	AnnotationLevel string `json:"annotation_level"`
	Message         string `json:"message"`
	Title           string `json:"title"`
	RawDetails      string `json:"raw_details"`
}

const (
	path           = "/hooks"
	tokenDurations = 60
)

var (
	hookSecret = os.Getenv("HOOK_SECRET")
)

func main() {

	http.HandleFunc(path, webhookHandler)
	http.ListenAndServe(":4444", nil)
}

func webhookHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("hook secret:", hookSecret)
	hook, _ := github.New(github.Options.Secret(hookSecret))
	payload, err := hook.Parse(r, github.PullRequestEvent, github.PushEvent)
	if err != nil {
		if err == github.ErrEventNotFound {
			fmt.Println("Event Not Registered:", err)
		}
	}
	switch payload.(type) {
	case github.PullRequestPayload:
		pullRequest := payload.(github.PullRequestPayload)
		//fmt.Printf("%+v", pullRequest)
		accessToken := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.StandardClaims{
			IssuedAt:  time.Now().Unix(),
			ExpiresAt: time.Now().Add(tokenDurations).Unix(),
			Issuer:    "17332",
		})
		token, err := accessToken.SignedString(privateKey())
		authenticate(token)
		checkRun := makeInProgressChecks(pullRequest.PullRequest.Head.Sha)
		inProgressPayload, err := json.Marshal(checkRun)
		fatal(err)
		req, err := http.NewRequest(
			"POST",
			"https://api.github.com/repos/alirezatjk/checks/check-runs",
			bytes.NewBuffer(inProgressPayload))
		fatal(err)
		req.Header.Set("Accept", "application/vnd.github.antiope-preview+json")
		println(token)
		client := &http.Client{}
		resp, err := client.Do(req)
		fatal(err)
		body, err := ioutil.ReadAll(resp.Body)
		defer resp.Body.Close()
		fatal(err)
		fmt.Println("Response: ", string(body))
	case github.PushPayload:
		moo := ""
		fmt.Printf("%+v", moo)
	}
}

//mooooooooooooooooooooooo
func createCheckRun(head string) CreateCheckRunPayload {
	checkRun := CreateCheckRunPayload{
		Name:        "First check run test",
		HeadSha:     head,
		DetailsURL:  "",
		ExternalID:  "12564",
		Status:      "in_progress",
		StartedAt:   time.Now().Format("2006-01-02T15:04:05Z07:00"),
		CompletedAt: time.Now().Format("2006-01-02T15:04:05Z07:00"),
		Conclusion:  "action_required",
		Output: Output{
			Title:   "First check title!",
			Summary: "Damn, if this works it's awesome!",
			Text:    "Blablablabla bla bla blabla bla blabla bla",
			Annotations: []Annotations{{
				Path:            "/static/css/stylesheet.css",
				StartLine:       2,
				EndLine:         3,
				StartColumn:     2,
				EndColumn:       3,
				AnnotationLevel: "warning",
				Message:         "YO this is annotation",
				Title:           "Well, this is title",
				RawDetails:      "This is raw details",
			}},
			Images: []Images{{
				Alt:      "Image alt",
				ImageURL: "https://dashboard.mielse.com/static/images/logo.png",
				Caption:  "Caption lel lel",
			}},
		},
		Actions: Actions{
			Description: "Action description",
			Identifier:  "Action Identifier",
			Label:       "Action Label",
		},
	}
	return checkRun
}

func makeInProgressChecks(head string) map[string]interface{} {
	inProgressCheck := map[string]interface{}{
		"name":        "mighty_readme",
		"head_sha":    head,
		"status":      "in_progress",
		"external_id": "42",
		"started_at":  time.Now().Format("2006-01-02T15:04:05Z07:00"),
		"output": map[string]string{
			"title":   "Mighty Readme report",
			"summary": "",
			"text":    "",
		},
	}
	return inProgressCheck
}

func fatal(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func privateKey() *rsa.PrivateKey {
	secret, err := ioutil.ReadFile("secret.pem")
	fatal(err)
	signedSecret, err := jwt.ParseRSAPrivateKeyFromPEM(secret)
	fatal(err)
	return signedSecret
}

func authenticate(token string) string {
	req, err := http.NewRequest(
		"POST",
		"https://api.github.com/app/installations/:installation_id/access_tokens",
		nil)
	fatal(err)
	req.Header.Set("Accept", "application/vnd.github.machine-man-preview+json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	client := &http.Client{}
	resp, err := client.Do(req)
	fatal(err)
	body, err := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()
	fatal(err)
	println(body)
	return string(body)
}
