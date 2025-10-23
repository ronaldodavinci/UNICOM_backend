package utils

import (
	"regexp"
	"strings"
)

// ExtractHashtags returns hashtags from text but drops any hashtag
// that contains profanity. Example: "#เหี้ยดี #ok" -> ["#ok"].
func ExtractHashtags(text string) []string {
	re := regexp.MustCompile(`#([\p{L}\p{M}0-9_]+)`)
	tags := re.FindAllString(text, -1)
	if len(tags) == 0 {
		return tags
	}
	out := make([]string, 0, len(tags))
	for _, tag := range tags {
		raw := strings.TrimPrefix(tag, "#")
		masked := MaskProfanity(raw)
		if masked != raw {
			// tag contains profanity -> drop
			continue
		}
		out = append(out, tag)
	}
	return out
}
