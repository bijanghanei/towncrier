package filter

import (
	"strings"
	"towncrier/publisher/external/models"
)

func FilterTweet(tweet models.Tweet, keywords []string) bool {
	for _, keyword := range(keywords) {
		if !strings.Contains(tweet.Text, keyword) {
			return false
		}
	}
	return true
}