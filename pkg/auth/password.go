// Package auth provides authentication utilities including password hashing and JWT token management.
package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"

	"golang.org/x/crypto/argon2"
)

// Password hashing errors.
var (
	// ErrInvalidHash is returned when the password hash format is invalid or malformed.
	ErrInvalidHash = errors.New("invalid password hash format")
	// ErrIncompatibleVersion is returned when the Argon2 version in the hash doesn't match the current version.
	ErrIncompatibleVersion = errors.New("incompatible argon2 version")
)

// Argon2id parameters following OWASP recommendations.
const (
	argon2Memory            = 64 * 1024 // 64 MB
	argon2Iterations        = 3
	argon2Parallelism       = 2
	argon2SaltLength        = 16
	argon2KeyLength         = 32
	argon2EncodedHashFields = 6 // Number of fields in PHC format: $algorithm$version$params$salt$hash
)

// hashParams holds the decoded parameters from a PHC-formatted Argon2id hash.
type hashParams struct {
	memory      uint32
	iterations  uint32
	parallelism uint8
	salt        []byte
	hash        []byte
}

// HashPassword generates an Argon2id hash of the given password.
// It returns the encoded hash string in PHC format ($argon2id$v=...$m=...,t=...,p=...$salt$hash),
// or an error if salt generation fails.
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

// VerifyPassword checks if the provided password matches the Argon2id encoded hash.
// It returns true if the password is correct, false otherwise.
// Returns an error if the hash format is invalid or incompatible.
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
		argon2KeyLength,
	)

	return subtle.ConstantTimeCompare(params.hash, otherHash) == 1, nil
}

// decodeHash parses a PHC-formatted Argon2id hash string and returns the parameters.
func decodeHash(encodedHash string) (*hashParams, error) {
	vals := strings.Split(encodedHash, "$")

	if len(vals) != argon2EncodedHashFields {
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

// HashToken generates a SHA-256 hash of the token for secure storage in the database.
// The original token should never be stored directly; only this hash should be persisted.
// This allows token verification without exposing the actual token value.
func HashToken(token string) string {
	hash := sha256.Sum256([]byte(token))

	return hex.EncodeToString(hash[:])
}
