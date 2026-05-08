package config

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
)

type Token struct {
	Token        string `json:"token"`
	RefreshToken string `json:"refresh_token,omitempty"`
	Roles        string `json:"roles,omitempty"`
}

func LoadToken(path string) (Token, error) {
	var tok Token
	b, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return tok, nil
		}
		return tok, err
	}
	if err := json.Unmarshal(b, &tok); err != nil {
		return tok, err
	}
	return tok, nil
}

func SaveToken(path string, tok Token) error {
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return err
	}
	b, err := json.MarshalIndent(tok, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, append(b, '\n'), 0600)
}
