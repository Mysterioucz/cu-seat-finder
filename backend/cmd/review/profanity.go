package main

import "strings"

// ponytail: hardcoded wordlist substring match, not a real moderation model.
// Upgrade path: pluggable filter (e.g. a hosted moderation API) if false
// positives/negatives on real reviews become a problem.
var profaneWords = []string{
	// English
	"fuck", "shit", "bitch", "asshole", "damn", "bastard",
	// Thai
	"เหี้ย", "สัส", "ควย", "เย็ด", "ไอสัตว์", "แม่ง",
}

func containsProfanity(text string) bool {
	lower := strings.ToLower(text)
	for _, w := range profaneWords {
		if strings.Contains(lower, strings.ToLower(w)) {
			return true
		}
	}
	return false
}

var allowedTags = map[string]bool{
	"quiet": true, "noisy": true, "cold": true, "hot": true,
	"strong_wifi": true, "weak_wifi": true, "clean": true,
	"many_outlets": true, "crowded": true,
}

func validTags(tags []string) bool {
	for _, t := range tags {
		if !allowedTags[t] {
			return false
		}
	}
	return true
}
