package main

import (
	"strings"
)

func icontains(needle string, haystack []string) bool {
	needle = strings.ToLower(needle)
	for _, val := range haystack {
		if needle == strings.ToLower(val) {
			return true
		}
	}
	return false
}
