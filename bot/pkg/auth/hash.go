package auth

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"github.com/google/uuid"
	"golang.org/x/crypto/argon2"
	"strings"
	"sync"
)

type HashHelper struct {
	format  string
	version int
	time    uint32
	memory  uint32
	keyLen  uint32
	saltLen uint32
	threads uint8
}

var once sync.Once
var hashHelperInstance *HashHelper

func NewUUID() string {
	return uuid.NewString()
}

func (h *HashHelper) Hash(plain string) (string, error) {
	salt := make([]byte, h.saltLen)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}
	hash := argon2.IDKey([]byte(plain), salt, h.time, h.memory, h.threads, h.keyLen)
	return fmt.Sprintf(
		h.format,
		h.version,
		h.memory,
		h.time,
		h.threads,
		base64.RawStdEncoding.EncodeToString(salt),
		base64.RawStdEncoding.EncodeToString(hash),
	), nil
}

func (h *HashHelper) Verify(plain string, hash string) (bool, error) {
	hashParts := strings.Split(hash, "$")
	var memory, time uint32
	var threads uint8
	_, err := fmt.Sscanf(hashParts[3], "m=%d,t=%d,p=%d", &memory, &time, &threads)
	if err != nil {
		return false, err
	}
	salt, err := base64.RawStdEncoding.DecodeString(hashParts[4])
	if err != nil {
		return false, err
	}
	decodedHash, err := base64.RawStdEncoding.DecodeString(hashParts[5])
	if err != nil {
		return false, err
	}
	compareHash := argon2.IDKey([]byte(plain), salt, time, memory, threads, uint32(len(decodedHash)))
	return subtle.ConstantTimeCompare(decodedHash, compareHash) == 1, nil
}

func GetHashHelperInstance() *HashHelper {
	if hashHelperInstance == nil {
		once.Do(func() {
			hashHelperInstance = &HashHelper{
				format:  "$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
				version: argon2.Version,
				time:    1,
				memory:  64 * 1024,
				keyLen:  32,
				saltLen: 16,
				threads: 2,
			}
		})
	}
	return hashHelperInstance
}
