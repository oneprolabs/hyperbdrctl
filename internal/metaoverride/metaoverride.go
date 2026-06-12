package metaoverride

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
)

type pathToken struct {
	key      string
	hasIndex bool
	index    int
}

func ReadObjectFile(path string, objectLabel string, forbiddenTopKeys map[string]string) (map[string]interface{}, error) {
	body, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	body = bytes.TrimPrefix(body, []byte{0xEF, 0xBB, 0xBF})

	var raw interface{}
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, err
	}

	switch typed := raw.(type) {
	case []interface{}:
		return nil, fmt.Errorf("file must contain single %s", objectLabel)
	case map[string]interface{}:
		for key, wrapperLabel := range forbiddenTopKeys {
			if _, ok := typed[key]; ok {
				return nil, fmt.Errorf("file must contain %s, not %s", objectLabel, wrapperLabel)
			}
		}
		return typed, nil
	default:
		return nil, fmt.Errorf("file must contain %s", objectLabel)
	}
}

func SplitAssignment(raw string) (string, string, error) {
	if raw == "" {
		return "", "", fmt.Errorf("assignment must be in path=value format")
	}
	idx := strings.Index(raw, "=")
	if idx <= 0 {
		return "", "", fmt.Errorf("assignment must be in path=value format")
	}
	return raw[:idx], raw[idx+1:], nil
}

func InferValue(raw string) interface{} {
	switch raw {
	case "true":
		return true
	case "false":
		return false
	}
	if intValue, err := strconv.Atoi(raw); err == nil {
		return intValue
	}
	return raw
}

func ApplyPathValue(metadata map[string]interface{}, path string, value interface{}) error {
	tokens, err := parsePath(path)
	if err != nil {
		return err
	}
	var current interface{} = metadata
	for i, token := range tokens {
		last := i == len(tokens)-1
		switch container := current.(type) {
		case map[string]interface{}:
			if !token.hasIndex {
				if last {
					container[token.key] = value
					continue
				}
				next, ok := container[token.key]
				if !ok {
					if tokens[i+1].hasIndex {
						return fmt.Errorf("path %s requires existing array at %s", path, token.key)
					}
					next = map[string]interface{}{}
					container[token.key] = next
				}
				current = next
				continue
			}
			next, ok := container[token.key]
			if !ok {
				return fmt.Errorf("path %s requires existing array at %s", path, token.key)
			}
			slice, ok := next.([]interface{})
			if !ok {
				return fmt.Errorf("path %s expects array at %s", path, token.key)
			}
			if token.index < 0 || token.index >= len(slice) {
				return fmt.Errorf("path %s index %d out of range", path, token.index)
			}
			if last {
				slice[token.index] = value
				container[token.key] = slice
				continue
			}
			current = slice[token.index]
		case []interface{}:
			if token.index < 0 || token.index >= len(container) {
				return fmt.Errorf("path %s index %d out of range", path, token.index)
			}
			if last {
				container[token.index] = value
				continue
			}
			current = container[token.index]
		default:
			return fmt.Errorf("path %s cannot descend into non-container value", path)
		}
	}
	return nil
}

func parsePath(path string) ([]pathToken, error) {
	if strings.TrimSpace(path) == "" {
		return nil, fmt.Errorf("path is required")
	}
	parts := strings.Split(path, ".")
	tokens := make([]pathToken, 0, len(parts))
	for _, part := range parts {
		if part == "" {
			return nil, fmt.Errorf("invalid path %q", path)
		}
		token, err := parsePathToken(part, path)
		if err != nil {
			return nil, err
		}
		tokens = append(tokens, token)
	}
	return tokens, nil
}

func parsePathToken(part, fullPath string) (pathToken, error) {
	if !strings.Contains(part, "[") {
		return pathToken{key: part}, nil
	}
	if !strings.HasSuffix(part, "]") {
		return pathToken{}, fmt.Errorf("invalid path %q", fullPath)
	}
	openIdx := strings.Index(part, "[")
	if openIdx <= 0 || strings.Count(part, "[") != 1 || strings.Count(part, "]") != 1 {
		return pathToken{}, fmt.Errorf("invalid path %q", fullPath)
	}
	key := part[:openIdx]
	indexText := part[openIdx+1 : len(part)-1]
	index, err := strconv.Atoi(indexText)
	if err != nil || index < 0 {
		return pathToken{}, fmt.Errorf("invalid path %q", fullPath)
	}
	return pathToken{key: key, hasIndex: true, index: index}, nil
}
