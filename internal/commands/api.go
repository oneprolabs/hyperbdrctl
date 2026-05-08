package commands

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
)

func runAPI(ctx *context, args []string) error {
	if len(args) == 0 {
		return errUnknown("api", "")
	}
	switch args[0] {
	case "request":
		return runAPIRequest(ctx, args[1:])
	default:
		return errUnknown("api", args[0])
	}
}

func runAPIRequest(ctx *context, args []string) error {
	fs := newFlagSet("api request")
	method := fs.String("method", "GET", "")
	path := fs.String("path", "", "")
	file := fs.String("file", "", "")
	inlineBody := fs.String("body", "", "")
	queryFlags := &keyValueFlags{}
	headerFlags := &keyValueFlags{}
	fs.Var(queryFlags, "query", "")
	fs.Var(headerFlags, "header", "")
	q := queryFromPairs()
	if err := parseQueryFlagsInto(fs, args, q); err != nil {
		return err
	}
	normalizedMethod, err := normalizeAPIMethod(*method)
	if err != nil {
		return err
	}
	if err := validateAPIPath(*path); err != nil {
		return err
	}
	if err := addKeyValueQuery(q, queryFlags.values); err != nil {
		return err
	}
	headers, err := apiHeaders(headerFlags.values)
	if err != nil {
		return err
	}
	body, err := apiRequestBody(*file, *inlineBody)
	if err != nil {
		return err
	}
	c, err := ensureClient(ctx)
	if err != nil {
		return err
	}
	resp, err := c.Request(normalizedMethod, *path, q, body, headers)
	if err != nil {
		return err
	}
	return writeResponse(ctx, resp, "", nil)
}

func normalizeAPIMethod(method string) (string, error) {
	method = strings.ToUpper(strings.TrimSpace(method))
	switch method {
	case "GET", "POST", "PUT", "PATCH", "DELETE":
		return method, nil
	default:
		return "", fmt.Errorf("method must be one of GET, POST, PUT, PATCH, DELETE")
	}
}

func validateAPIPath(path string) error {
	if path == "" {
		return fmt.Errorf("path is required")
	}
	if !strings.HasPrefix(path, "/") || strings.HasPrefix(path, "//") {
		return fmt.Errorf("path must be a host-relative path starting with /")
	}
	u, err := url.Parse(path)
	if err != nil {
		return err
	}
	if u.Scheme != "" || u.Host != "" {
		return fmt.Errorf("path must not be an absolute URL")
	}
	return nil
}

func apiRequestBody(file, inlineBody string) (interface{}, error) {
	if file != "" && inlineBody != "" {
		return nil, fmt.Errorf("file and body are mutually exclusive")
	}
	if file == "" && inlineBody == "" {
		return nil, nil
	}
	var raw []byte
	var err error
	if file != "" {
		raw, err = os.ReadFile(file)
		if err != nil {
			return nil, err
		}
	} else {
		raw = []byte(inlineBody)
	}
	var body interface{}
	if err := json.Unmarshal(raw, &body); err != nil {
		return nil, err
	}
	return body, nil
}

func addKeyValueQuery(q url.Values, values []string) error {
	for _, value := range values {
		key, item, ok := strings.Cut(value, "=")
		if !ok || key == "" {
			return fmt.Errorf("query must be key=value")
		}
		q.Add(key, item)
	}
	return nil
}

func apiHeaders(values []string) (http.Header, error) {
	headers := http.Header{}
	for _, value := range values {
		key, item, ok := strings.Cut(value, "=")
		if !ok || key == "" {
			return nil, fmt.Errorf("header must be key=value")
		}
		if protectedAPIHeader(key) {
			return nil, fmt.Errorf("header %s is managed by hyperbdrctl and cannot be overridden", key)
		}
		headers.Add(key, item)
	}
	return headers, nil
}

func protectedAPIHeader(key string) bool {
	switch http.CanonicalHeaderKey(key) {
	case "X-Auth-Token", "X-Scene", "X-Lang", "Accept", "Content-Type":
		return true
	default:
		return false
	}
}

type keyValueFlags struct {
	values []string
}

func (f *keyValueFlags) Set(value string) error {
	f.values = append(f.values, value)
	return nil
}

func (f *keyValueFlags) String() string {
	return strings.Join(f.values, ",")
}
