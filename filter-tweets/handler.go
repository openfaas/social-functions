package function

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
)

// Handle a serverless request
func Handle(req []byte) string {

	currentTweet := tweet{}
	unmarshalErr := json.Unmarshal(req, &currentTweet)

	if unmarshalErr != nil {
		return fmt.Sprintf("Unable to unmarshal event: %s", unmarshalErr.Error())
	}

	if strings.Contains(currentTweet.Text, "RT") || currentTweet.Text == "alexellisuk_bot" || currentTweet.Username == "colorisebot" {
		return fmt.Sprintf("Filtered the tweet out")
	}

	client := http.Client{}
	slackURL := os.Getenv("slack_url")

	slackMsg := slackMessage{
		Text:     "@" + currentTweet.Username + ": " + currentTweet.Text + " (via " + currentTweet.Link + ")",
		Username: "@" + currentTweet.Username,
	}

	bodyBytes, _ := json.Marshal(slackMsg)
	httpReq, _ := http.NewRequest(http.MethodPost, slackURL, bytes.NewReader(bodyBytes))
	res, reqErr := client.Do(httpReq)
	if reqErr != nil {
		fmt.Fprintf(os.Stderr, "reqErr: %s", reqErr)
		os.Exit(1)
	}

	if res.Body != nil {
		defer res.Body.Close()
	}

	return fmt.Sprintf("Tweet sent, with statusCode: %d", res.StatusCode)
}

// tweet in following format from IFTTT:
// { "text": "<<<{{Text}}>>>", "username": "<<<{{UserName}}>>>", "link": "<<<{{LinkToTweet}}>>>" }
type tweet struct {
	Text     string `json:"text"`
	Username string `json:"username"`
	Link     string `json:"link"`
}

type slackMessage struct {
	Text     string `json:"text"`
	Username string `json:"username"`
}
