package domain

import (
	"regexp"
	"strings"
)

func normalize(input string) string {
	input = strings.TrimSpace(input)
	input = strings.ToLower(input)

	reg := regexp.MustCompile(`[^a-z0-9]+`)
	input = reg.ReplaceAllString(input, "_")
	input = strings.Trim(input, "_")

	return input
}

func NormalizeMap(input map[string]string) map[string]string {
	result := make(map[string]string)

	for k, v := range input {
		key := strings.ToLower(strings.TrimSpace(k))
		key = strings.ReplaceAll(key, " ", "_")

		result[key] = v
	}

	return result
}