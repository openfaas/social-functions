package function

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"strings"
)

// Handle a serverless request
func Handle(req []byte) string {
	currentTweet := tweet{}

	if err := json.Unmarshal(req, &currentTweet); err != nil {
		return fmt.Sprintf("Unable to unmarshal event: %s", err.Error())
	}

	if strings.Contains(currentTweet.Text, "RT") ||
		currentTweet.Text == "alexellisuk_bot" ||
		currentTweet.Username == "colorisebot" ||
		currentTweet.Username == "scmsFaAS" ||
		currentTweet.Username == "openfaas" {
		return "filtered the tweet out"
	}

	discordURL := readSecret("twitter-discord-webhook-url")
	discordMsg := discordMessage{
		Content:  "@" + currentTweet.Username + ": " + currentTweet.Text + " (via " + currentTweet.Link + ")",
		Username: "@" + currentTweet.Username,
	}

	bodyBytes, _ := json.Marshal(discordMsg)
	httpReq, err := http.NewRequest(http.MethodPost, discordURL, bytes.NewReader(bodyBytes))
	if err != nil {
		if err != nil {
			fmt.Fprintf(os.Stderr, "resErr: %s", err)
			os.Exit(1)
		}
	}

	httpReq.Header.Set("Content-Type", "application/json")

	res, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		fmt.Fprintf(os.Stderr, "resErr: %s", err)
		os.Exit(1)
	}

	if res.Body != nil {
		defer res.Body.Close()
	}

	bodyRes, _ := ioutil.ReadAll(res.Body)

	if res.StatusCode != http.StatusAccepted &&
		res.StatusCode != http.StatusOK &&
		res.StatusCode != http.StatusNoContent {
		fmt.Fprintf(os.Stderr, "unexpected status code: %d, body: %s", res.StatusCode, string(bodyRes))
		os.Exit(1)
	}

	return fmt.Sprintf("tweet forwarded [%d]", res.StatusCode)
}

// tweet in following format from IFTTT:
// { "text": "<<<{{Text}}>>>", "username": "<<<{{UserName}}>>>", "link": "<<<{{LinkToTweet}}>>>" }
type tweet struct {
	Text     string `json:"text"`
	Username string `json:"username"`
	Link     string `json:"link"`
}

type discordMessage struct {
	Content  string `json:"content"`
	Username string `json:"username"`
}

func readSecret(name string) string {
	res, err := ioutil.ReadFile(path.Join("/var/openfaas/secrets/", name))
	if err != nil {
		os.Stderr.Write([]byte(err.Error()))
		os.Exit(1)
	}
	return strings.TrimSpace(string(res))
}
