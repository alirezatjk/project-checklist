package main

import (
	"bytes"
	"crypto/rsa"
	"encoding/json"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"gopkg.in/go-playground/webhooks.v5/github"
	"io/ioutil"
	"net/http"
	"os"
	"time"
)

type GithubToken struct {
	Token     string `json:"token"`
	ExpiresAt string `json:"expires_at"`
}

//
type CheckRunStatusPayload struct {
	Name        string    `json:"name"`
	HeadSha     string    `json:"head_sha"`
	DetailsURL  string    `json:"details_url"`
	ExternalID  string    `json:"external_id,omitempty"`
	Status      string    `json:"status"`
	StartedAt   string    `json:"started_at"`
	Conclusion  string    `json:"conclusion,omitempty"`
	CompletedAt string    `json:"completed_at,omitempty"`
	Output      Output    `json:"output"`
	Actions     []Actions `json:"actions,omitempty"`
}

type Actions struct {
	Label       string `json:"label"`
	Description string `json:"description"`
	Identifier  string `json:"identifier"`
}

type Output struct {
	Title       string        `json:"title"`
	Summary     string        `json:"summary"`
	Text        string        `json:"text,omitempty"`
	Annotations []Annotations `json:"annotations,omitempty"`
	Images      []Images      `json:"images,omitempty"`
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
	hook, _ := github.New(github.Options.Secret(hookSecret))
	payload, err := hook.Parse(r, github.PullRequestEvent, github.PushEvent)
	if err != nil {
		fmt.Println(err)
	}
	switch payload.(type) {
	case github.PullRequestPayload:
		pullRequest := payload.(github.PullRequestPayload)
		token := createAuthenticationToken()
		checkRun, err := createInProgressChecks(pullRequest.PullRequest.Head.Sha)
		if err != nil {
			fmt.Println(err)
		}
		inProgressCheckRun := sendCheckRunRequest(checkRun, token)
		fmt.Println("Response: ", string(inProgressCheckRun))
	}
}

func sendCheckRunRequest(payload []byte, token string) []byte {
	req, err := http.NewRequest(
		"POST",
		"https://api.github.com/repos/alirezatjk/checks/check-runs",
		bytes.NewBuffer(payload))
	if err != nil {
		fmt.Println(err)
	}
	req.Header.Set("Accept", "application/vnd.github.antiope-preview+json")
	req.Header.Set("Authorization", fmt.Sprintf("token %s", authenticate(token)))
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err)
	}
	defer resp.Body.Close()
	return body
}

func createInProgressChecks(head string) ([]byte, error) {
	images := []Images{
		{
			Alt:      "Becoming one in a blue environment yo",
			ImageURL: "http://fertilitymatters.ca/wp-content/uploads/2018/01/Zen.jpg",
			Caption:  "Becoming one in a blue environment",
		},
		{
			Alt:      "Becoming one in a yellowish environment yo",
			ImageURL: "https://img.purch.com/h/1400/aHR0cDovL3d3dy5saXZlc2NpZW5jZS5jb20vaW1hZ2VzL2kvMDAwLzAwMi81MDUvb3JpZ2luYWwvMDgwOTAyLXplbi1tZWRpdGF0aW9uLTAyLmpwZw==",
			Caption:  "Becoming one in a yellowish environment",
		},
		{
			Alt:      "Becoming one in a blue environment again yo",
			ImageURL: "https://mylifemystuff.files.wordpress.com/2012/04/zen.jpg",
			Caption:  "Becoming one in a blue environment again",
		},
	}
	actions := []Actions{
		{
			Label:       "Kill Tabbat",
			Description: "Kill Tabbat to speed up process",
			Identifier:  "kill_tabbat",
		},
		{
			Label:       "Don't kill Tabbat",
			Description: "Don't kill Tabbat to slow down process",
			Identifier:  "dont_kill_tabbat",
		},
	}
	output := Output{
		Title:   "Testing Consciousness",
		Summary: "I am becoming one with the code",
		Text:    "All that we are, is the result of what we have thought. The mind is everything. What we think, we become.",
		Images:  images,
	}
	inProgressCheck := CheckRunStatusPayload{
		Name:       "Smoke Test",
		HeadSha:    head,
		Status:     "in_progress",
		ExternalID: "69696969",
		StartedAt:  time.Now().Format("2006-01-02T15:04:05Z07:00"),
		Output:     output,
		Actions:    actions,
	}

	return json.Marshal(inProgressCheck)
}

func getSignedSecret() (*rsa.PrivateKey, error) {
	secret, err := ioutil.ReadFile("secret.pem")
	if err != nil {
		return nil, err
	}
	signedSecret, err := jwt.ParseRSAPrivateKeyFromPEM(secret)
	if err != nil {
		return nil, err
	}
	return signedSecret, nil
}

func authenticate(accessToken string) string {
	req, err := http.NewRequest(
		"POST",
		"https://api.github.com/app/installations/427948/access_tokens",
		nil,
	)
	if err != nil {
		fmt.Println(err)
	}
	req.Header.Set("Accept", "application/vnd.github.machine-man-preview+json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", accessToken))
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
	}
	var token GithubToken
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err)
	}
	defer resp.Body.Close()
	err = json.Unmarshal(body, &token)
	if err != nil {
		fmt.Println(err)
	}
	return token.Token
}

func createAuthenticationToken() string {
	accessToken := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.StandardClaims{
		IssuedAt:  time.Now().Unix(),
		ExpiresAt: time.Now().Add(tokenDurations).Unix(),
		Issuer:    "17332",
	})
	privateKey, err := getSignedSecret()
	if err != nil {
		panic(fmt.Sprintf("Cannot get signed secret: %s", err))
	}
	token, err := accessToken.SignedString(privateKey)
	if err != nil {
		panic(fmt.Sprintf("Cannot create authentication token: %s", err))
	}
	return token
}
