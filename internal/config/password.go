package config

import (
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
)

const passwordEncodingPrefix = "enc:v1:"

func encodePassword(username, password string) (string, error) {
	if password == "" {
		return "", nil
	}
	key := []byte(username)
	if len(key) == 0 {
		return "", errors.New("username is required to encode password")
	}
	return passwordEncodingPrefix + hex.EncodeToString(xorPassword([]byte(password), key)), nil
}

func decodePassword(username, password string) (string, error) {
	if !strings.HasPrefix(password, passwordEncodingPrefix) {
		return password, nil
	}
	key := []byte(username)
	if len(key) == 0 {
		return "", errors.New("username is required to decode password")
	}
	encoded, err := hex.DecodeString(strings.TrimPrefix(password, passwordEncodingPrefix))
	if err != nil {
		return "", fmt.Errorf("decode stored password: %w", err)
	}
	return string(xorPassword(encoded, key)), nil
}

func xorPassword(value, key []byte) []byte {
	result := make([]byte, len(value))
	for i, b := range value {
		result[i] = b ^ key[i%len(key)]
	}
	return result
}
