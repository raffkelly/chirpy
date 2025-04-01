package auth

import "testing"

func TestEnterCorrectPassword(t *testing.T) {
	password := "WhatInTheFuck"
	hashedPassword, err := HashPassword(password)
	if err != nil {
		t.Fatalf("error creating hashed password")
	}
	entered := "WhatInTheFuck"
	err = CheckPasswordHash(hashedPassword, entered)
	if err != nil {
		t.Fatalf("error checking entered password against hashed password")
	}
}

func TestEnterIncorrectPassword(t *testing.T) {
	password := "WhatInTheFuck"
	hashedPassword, err := HashPassword(password)
	if err != nil {
		t.Fatalf("error creating hashed password")
	}
	entered := "WhatInTheFuck2"
	err = CheckPasswordHash(hashedPassword, entered)
	if err == nil {
		t.Fatalf("incorrect password passed hash check")
	}
}
