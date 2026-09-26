package main

import "testing"

func TestContainsProfanityEnglish(t *testing.T) {
	if !containsProfanity("this place is shit") {
		t.Error("expected profanity to be detected")
	}
	if !containsProfanity("SHIT is bad even uppercase") {
		t.Error("expected case-insensitive match")
	}
}

func TestContainsProfanityThai(t *testing.T) {
	if !containsProfanity("ห้องสมุดนี้ เหี้ย มาก") {
		t.Error("expected Thai profanity to be detected")
	}
}

func TestContainsProfanityCleanText(t *testing.T) {
	if containsProfanity("Great quiet library with strong wifi") {
		t.Error("clean text should not be flagged")
	}
}

func TestValidTags(t *testing.T) {
	if !validTags([]string{"quiet", "strong_wifi"}) {
		t.Error("expected known tags to be valid")
	}
	if validTags([]string{"quiet", "smelly"}) {
		t.Error("expected unknown tag to be invalid")
	}
	if !validTags(nil) {
		t.Error("empty tag list should be valid")
	}
}
