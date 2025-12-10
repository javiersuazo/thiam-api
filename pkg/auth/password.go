// Package auth provides authentication utilities.
package auth

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"

	"golang.org/x/crypto/argon2"
)

// Password hashing errors.
var (
	ErrInvalidHash         = errors.New("invalid password hash format")
	ErrIncompatibleVersion = errors.New("incompatible argon2 version")
)

// Argon2 configuration constants.
const (
	argon2Memory      = 64 * 1024
	argon2Iterations  = 3
	argon2Parallelism = 2
	argon2SaltLength  = 16
	argon2KeyLength   = 32
	argon2HashParts   = 6
)

// hashParams holds the decoded argon2 parameters.
type hashParams struct {
	memory      uint32
	iterations  uint32
	parallelism uint8
	salt        []byte
	hash        []byte
}

// HashPassword creates an Argon2id hash of the password.
func HashPassword(password string) (string, error) {
	salt := make([]byte, argon2SaltLength)

	if _, err := rand.Read(salt); err != nil {
		return "", fmt.Errorf("generating salt: %w", err)
	}

	hash := argon2.IDKey(
		[]byte(password),
		salt,
		argon2Iterations,
		argon2Memory,
		argon2Parallelism,
		argon2KeyLength,
	)

	b64Salt := base64.RawStdEncoding.EncodeToString(salt)
	b64Hash := base64.RawStdEncoding.EncodeToString(hash)

	return fmt.Sprintf(
		"$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2.Version,
		argon2Memory,
		argon2Iterations,
		argon2Parallelism,
		b64Salt,
		b64Hash,
	), nil
}

// VerifyPassword checks if the password matches the hash.
func VerifyPassword(password, encodedHash string) (bool, error) {
	params, err := decodeHash(encodedHash)
	if err != nil {
		return false, err
	}

	otherHash := argon2.IDKey(
		[]byte(password),
		params.salt,
		params.iterations,
		params.memory,
		params.parallelism,
		uint32(len(params.hash)),
	)

	return subtle.ConstantTimeCompare(params.hash, otherHash) == 1, nil
}

func decodeHash(encodedHash string) (*hashParams, error) {
	vals := strings.Split(encodedHash, "$")

	if len(vals) != argon2HashParts {
		return nil, ErrInvalidHash
	}

	var version int

	_, err := fmt.Sscanf(vals[2], "v=%d", &version)
	if err != nil {
		return nil, ErrInvalidHash
	}

	if version != argon2.Version {
		return nil, ErrIncompatibleVersion
	}

	params := &hashParams{}

	_, err = fmt.Sscanf(vals[3], "m=%d,t=%d,p=%d", &params.memory, &params.iterations, &params.parallelism)
	if err != nil {
		return nil, ErrInvalidHash
	}

	params.salt, err = base64.RawStdEncoding.DecodeString(vals[4])
	if err != nil {
		return nil, ErrInvalidHash
	}

	params.hash, err = base64.RawStdEncoding.DecodeString(vals[5])
	if err != nil {
		return nil, ErrInvalidHash
	}

	return params, nil
}
