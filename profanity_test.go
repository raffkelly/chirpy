package main

import "testing"

func TestRemoveProfanity(t *testing.T) {
	input := "This is a kerfuffle opinion I need to share."
	expected := "This is a **** opinion I need to share."

	// Assume RemoveProfanity is your function
	output := removeProfanity(input)

	if output != expected {
		t.Errorf("Expected '%s' but got '%s'", expected, output)
	}
}
