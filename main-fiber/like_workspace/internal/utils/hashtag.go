package utils

import (
	"regexp"
)

func ExtractHashtags(text string) []string {
	re := regexp.MustCompile(`#([\p{L}\p{M}0-9_]+)`)

	result := re.FindAllString(text, -1)

	return result
}
