package security

import "testing"

func TestHashAndVerifyPassword(t *testing.T) {
	password := "MyStrongPassword123!"

	hash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword returned error: %v", err)
	}

	if hash == password {
		t.Fatal("password must not be stored as plain text")
	}

	matched, err := VerifyPassword(password, hash)
	if err != nil {
		t.Fatalf("VerifyPassword returned error: %v", err)
	}

	if !matched {
		t.Fatal("correct password was rejected")
	}

	matched, err = VerifyPassword("WrongPassword123!", hash)
	if err != nil {
		t.Fatalf("VerifyPassword returned error: %v", err)
	}

	if matched {
		t.Fatal("incorrect password was accepted")
	}
}

func TestVerifyPasswordRejectsInvalidHash(t *testing.T) {
	_, err := VerifyPassword("password", "invalid-hash")

	if err == nil {
		t.Fatal("expected an error for invalid hash")
	}
}
