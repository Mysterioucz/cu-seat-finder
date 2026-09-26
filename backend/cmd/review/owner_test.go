package main

import "testing"

func TestCanDelete(t *testing.T) {
	if !canDelete("customer", "u1", "u1") || !canDelete("admin", "a", "u1") || canDelete("customer", "u2", "u1") {
		t.Fatal("ownership rule wrong")
	}
}
