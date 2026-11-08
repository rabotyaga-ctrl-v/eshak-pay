package security

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"

	"golang.org/x/crypto/argon2"

	"crypto/subtle"
	"strconv"
	"strings"
)

const (
	saltLength  = 16
	keyLength   = 32
	memory      = 64 * 1024
	iterations  = 3
	parallelism = 2
)

func HashPassword(password string) (string, error) {
	salt := make([]byte, saltLength)

	if _, err := rand.Read(salt); err != nil {
		return "", fmt.Errorf("generate password salt: %w", err)
	}

	hash := argon2.IDKey(
		[]byte(password),
		salt,
		iterations,
		memory,
		parallelism,
		keyLength,
	)

	encodedSalt := base64.RawStdEncoding.EncodeToString(salt)
	encodedHash := base64.RawStdEncoding.EncodeToString(hash)

	return fmt.Sprintf(
		"argon2id$%d$%d$%d$%s$%s",
		memory,
		iterations,
		parallelism,
		encodedSalt,
		encodedHash,
	), nil
}

func VerifyPassword(password, encoded string) (bool, error) {
	parts := strings.Split(encoded, "$")
	if len(parts) != 6 || parts[0] != "argon2id" {
		return false, fmt.Errorf("invalid password hash format")
	}

	memoryValue, err := strconv.ParseUint(parts[1], 10, 32)
	if err != nil || memoryValue < 8*1024 || memoryValue > 256*1024 {
		return false, fmt.Errorf("invalid Argon2 memory parameter")
	}

	iterationsValue, err := strconv.ParseUint(parts[2], 10, 32)
	if err != nil || iterationsValue < 1 || iterationsValue > 10 {
		return false, fmt.Errorf("invalid Argon2 iterations parameter")
	}

	parallelismValue, err := strconv.ParseUint(parts[3], 10, 8)
	if err != nil || parallelismValue < 1 || parallelismValue > 8 {
		return false, fmt.Errorf("invalid Argon2 parallelism parameter")
	}

	salt, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil || len(salt) < 16 || len(salt) > 64 {
		return false, fmt.Errorf("invalid password salt")
	}

	expectedHash, err := base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil || len(expectedHash) < 16 || len(expectedHash) > 64 {
		return false, fmt.Errorf("invalid password hash")
	}

	actualHash := argon2.IDKey(
		[]byte(password),
		salt,
		uint32(iterationsValue),
		uint32(memoryValue),
		uint8(parallelismValue),
		uint32(len(expectedHash)),
	)

	return subtle.ConstantTimeCompare(actualHash, expectedHash) == 1, nil
}
