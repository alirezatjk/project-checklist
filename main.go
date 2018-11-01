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

type GithubToken struct {
	Token     string `json:"token"`
	ExpiresAt string `json:"expires_at"`
}

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
	fatal(err)
	switch payload.(type) {
	case github.PullRequestPayload:
		pullRequest := payload.(github.PullRequestPayload)
		//fmt.Printf("%+v", pullRequest)
		token := createAuthenticationToken()
		checkRun := createInProgressChecks(pullRequest.PullRequest.Head.Sha)
		inProgressCheckRun := sendCheckRunRequest(checkRun, token)
		fmt.Println("Response: ", string(inProgressCheckRun))
	}
}

func sendCheckRunRequest(payload []byte, token string) []byte {
	req, err := http.NewRequest(
		"POST",
		"https://api.github.com/repos/alirezatjk/checks/check-runs",
		bytes.NewBuffer(payload))
	fatal(err)
	req.Header.Set("Accept", "application/vnd.github.antiope-preview+json")
	req.Header.Set("Authorization", fmt.Sprintf("token %s", authenticate(token)))
	client := &http.Client{}
	resp, err := client.Do(req)
	fatal(err)
	body, err := ioutil.ReadAll(resp.Body)
	fatal(err)
	defer resp.Body.Close()
	return body
}

func createInProgressChecks(head string) []byte {
	inProgressCheck := map[string]interface{}{
		"name":        "Smoke Test",
		"head_sha":    head,
		"status":      "in_progress",
		"external_id": "69698585",
		"started_at":  time.Now().Format("2006-01-02T15:04:05Z07:00"),
		"output": map[string]interface{}{
			"title":   "Testing Consciousness",
			"summary": "I am becoming one with the code",
			"text":    "All that we are, is the result of what we have thought. The mind is everything. What we think, we become",
			"images": []map[string]string{{
				"alt":       "Becoming one in a blue environment yo",
				"image_url": "http://fertilitymatters.ca/wp-content/uploads/2018/01/Zen.jpg",
				"caption":   "Becoming one in a blue environment",
			},
				{
					"alt":       "Becoming one in a yellowish environment yo",
					"image_url": "https://img.purch.com/h/1400/aHR0cDovL3d3dy5saXZlc2NpZW5jZS5jb20vaW1hZ2VzL2kvMDAwLzAwMi81MDUvb3JpZ2luYWwvMDgwOTAyLXplbi1tZWRpdGF0aW9uLTAyLmpwZw==",
					"caption":   "Becoming one in a yellowish environment",
				},
				{
					"alt":       "Becoming one in a blue environment again yo",
					"image_url": "https://mylifemystuff.files.wordpress.com/2012/04/zen.jpg",
					"caption":   "Becoming one in a blue environment again",
				},
			},
		},
		"actions": []map[string]string{
			{
				"label":       "Kill Tabbat",
				"description": "Kill Tabbat to speed up process",
				"identifier":  "kill_tabbat",
			},
			{
				"label":       "Don't kill Tabbat",
				"description": "Don't kill Tabbat to kinda speed up process",
				"identifier":  "dont_kill_tabbat",
			},
		},
	}
	payload, err := json.Marshal(inProgressCheck)
	fatal(err)
	return payload
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

func authenticate(accessToken string) string {
	req, err := http.NewRequest(
		"POST",
		"https://api.github.com/app/installations/427948/access_tokens",
		nil,
	)
	fatal(err)
	req.Header.Set("Accept", "application/vnd.github.machine-man-preview+json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", accessToken))
	client := &http.Client{}
	resp, err := client.Do(req)
	fatal(err)
	var token GithubToken
	body, err := ioutil.ReadAll(resp.Body)
	fatal(err)
	defer resp.Body.Close()
	err = json.Unmarshal(body, &token)
	fatal(err)
	println(body)
	return token.Token
}

func createAuthenticationToken() string {
	accessToken := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.StandardClaims{
		IssuedAt:  time.Now().Unix(),
		ExpiresAt: time.Now().Add(tokenDurations).Unix(),
		Issuer:    "17332",
	})
	token, err := accessToken.SignedString(privateKey())
	fatal(err)
	return token
}
