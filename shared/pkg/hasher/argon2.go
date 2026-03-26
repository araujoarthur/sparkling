// argon2.go implements Hasher using the argon2id algorithm.
package hasher

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"strings"

	"golang.org/x/crypto/argon2"
)

// Argon2Hasher implements Hasher using argon2id.
// The hash output includes the salt and parameters so it is self-contained
// and can be verified without storing them separately.
type Argon2Hasher struct {
	memory      uint32
	iterations  uint32
	parallelism uint8
	saltLength  uint32
	keyLength   uint32
}

// NewArgon2Hasher constructs an Argon2Hasher with recommended parameters.
func NewArgon2Hasher() *Argon2Hasher {
	return &Argon2Hasher{
		memory:      64 * 1024, // 64MB
		iterations:  3,
		parallelism: 2,
		saltLength:  16,
		keyLength:   32,
	}
}

// Hash returns an argon2id hash of the plaintext.
// The output format is: $argon2id$v=19$m=65536,t=3,p=2$<salt>$<hash>
func (h *Argon2Hasher) Hash(plaintext string) (string, error) {
	salt := make([]byte, h.saltLength)
	if _, err := rand.Read(salt); err != nil {
		return "", fmt.Errorf("Argon2Hasher.Hash [salt]: %w", err)
	}

	hash := argon2.IDKey(
		[]byte(plaintext),
		salt,
		h.iterations,
		h.memory,
		h.parallelism,
		h.keyLength,
	)

	encoded := fmt.Sprintf(
		"$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2.Version,
		h.memory,
		h.iterations,
		h.parallelism,
		base64.RawStdEncoding.EncodeToString(salt),
		base64.RawStdEncoding.EncodeToString(hash),
	)

	return encoded, nil
}

// Verify checks whether the plaintext matches the argon2id hash.
// It extracts the parameters and salt from the stored hash string.
func (h *Argon2Hasher) Verify(plaintext, hash string) (bool, error) {
	parts := strings.Split(hash, "$")
	if len(parts) != 6 {
		return false, fmt.Errorf("Argon2Hasher.Verify: invalid hash format")
	}

	var memory, iterations uint32
	var parallelism uint8
	_, err := fmt.Sscanf(parts[3], "m=%d,t=%d,p=%d", &memory, &iterations, &parallelism)
	if err != nil {
		return false, fmt.Errorf("Argon2Hasher.Verify [params]: %w", err)
	}

	salt, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		return false, fmt.Errorf("Argon2Hasher.Verify [salt]: %w", err)
	}

	storedHash, err := base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil {
		return false, fmt.Errorf("Argon2Hasher.Verify [hash]: %w", err)
	}

	computedHash := argon2.IDKey(
		[]byte(plaintext),
		salt,
		iterations,
		memory,
		parallelism,
		uint32(len(storedHash)),
	)

	if subtle.ConstantTimeCompare(computedHash, storedHash) != 1 {
		return false, nil
	}

	return true, nil
}
