package client

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"hyperbdr-client/internal/config"
)

type Client struct {
	cfg        config.Resolved
	httpClient *http.Client
	token      string
	debugOut   io.Writer
}

var sensitiveDebugKeys = map[string]bool{
	"password":       true,
	"user_password":  true,
	"reset_password": true,
	"auth_cert":      true,
	"token":          true,
	"refresh_token":  true,
}

type APIResponse struct {
	Code    string                 `json:"code,omitempty"`
	Data    interface{}            `json:"data,omitempty"`
	Error   interface{}            `json:"error,omitempty"`
	Title   string                 `json:"title,omitempty"`
	TraceID string                 `json:"trace_id,omitempty"`
	Raw     map[string]interface{} `json:"-"`
}

type HTTPError struct {
	StatusCode int
	Message    string
}

type APIError struct {
	Code    string
	Message string
}

func (e HTTPError) Error() string {
	if e.Message == "" {
		return fmt.Sprintf("http %d", e.StatusCode)
	}
	return fmt.Sprintf("http %d: %s", e.StatusCode, e.Message)
}

func (e APIError) Error() string {
	if e.Message == "" {
		return fmt.Sprintf("api %s", e.Code)
	}
	return fmt.Sprintf("api %s: %s", e.Code, e.Message)
}

func New(cfg config.Resolved) (*Client, error) {
	return NewWithDebug(cfg, io.Discard)
}

func NewWithDebug(cfg config.Resolved, debugOut io.Writer) (*Client, error) {
	transport := http.DefaultTransport.(*http.Transport).Clone()
	if cfg.Insecure {
		transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	}
	tok, err := config.LoadToken(cfg.CachePath)
	if err != nil {
		return nil, err
	}
	return &Client{
		cfg: cfg,
		httpClient: &http.Client{
			Timeout:   60 * time.Second,
			Transport: transport,
		},
		token:    tok.Token,
		debugOut: debugOut,
	}, nil
}

func (c *Client) Login() (config.Token, error) {
	if c.cfg.Host == "" {
		return config.Token{}, fmt.Errorf("host is required")
	}
	if c.cfg.Username == "" {
		return config.Token{}, fmt.Errorf("username is required")
	}
	if c.cfg.Password == "" {
		return config.Token{}, fmt.Errorf("password is required")
	}
	body := map[string]string{"username": c.cfg.Username, "password": c.cfg.Password}
	resp, err := c.do("POST", "/api/v2/loginNoImageCaptcha", nil, body, false)
	if err != nil {
		return config.Token{}, err
	}
	data, _ := resp.Data.(map[string]interface{})
	tok := config.Token{
		Token:        stringValue(data["token"]),
		RefreshToken: stringValue(data["refresh_token"]),
		Roles:        stringValue(data["roles"]),
	}
	if tok.Token == "" {
		return config.Token{}, fmt.Errorf("login response did not include token")
	}
	if err := config.SaveToken(c.cfg.CachePath, tok); err != nil {
		return config.Token{}, err
	}
	c.token = tok.Token
	return tok, nil
}

func (c *Client) Get(path string, q url.Values) (APIResponse, error) {
	return c.Request("GET", path, q, nil, nil)
}

func (c *Client) GetRaw(path string, q url.Values) ([]byte, error) {
	return c.requestRaw("GET", path, q, nil, nil)
}

func (c *Client) Post(path string, body interface{}) (APIResponse, error) {
	return c.Request("POST", path, nil, body, nil)
}

func (c *Client) Delete(path string, body interface{}) (APIResponse, error) {
	return c.Request("DELETE", path, nil, body, nil)
}

func (c *Client) Request(method, path string, q url.Values, body interface{}, headers http.Header) (APIResponse, error) {
	if err := c.ensureAuthenticated(); err != nil {
		return APIResponse{}, err
	}
	method = strings.ToUpper(method)
	resp, err := c.doWithHeaders(method, path, q, body, true, headers)
	if isUnauthorized(err) {
		if _, loginErr := c.Login(); loginErr != nil {
			return APIResponse{}, err
		}
		return c.doWithHeaders(method, path, q, body, true, headers)
	}
	return resp, err
}

func (c *Client) requestRaw(method, path string, q url.Values, body interface{}, headers http.Header) ([]byte, error) {
	if err := c.ensureAuthenticated(); err != nil {
		return nil, err
	}
	method = strings.ToUpper(method)
	respBody, err := c.doRawWithHeaders(method, path, q, body, true, headers)
	if isUnauthorized(err) {
		if _, loginErr := c.Login(); loginErr != nil {
			return nil, err
		}
		return c.doRawWithHeaders(method, path, q, body, true, headers)
	}
	return respBody, err
}

func (c *Client) ensureAuthenticated() error {
	if c.token != "" {
		return nil
	}
	if c.cfg.Username == "" || c.cfg.Password == "" {
		return nil
	}
	_, err := c.Login()
	return err
}

func (c *Client) do(method, path string, q url.Values, body interface{}, auth bool) (APIResponse, error) {
	return c.doWithHeaders(method, path, q, body, auth, nil)
}

func (c *Client) doRawWithHeaders(method, path string, q url.Values, body interface{}, auth bool, headers http.Header) ([]byte, error) {
	u, err := c.url(path, q)
	if err != nil {
		return nil, err
	}
	var r io.Reader
	debugBody := ""
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		r = bytes.NewReader(b)
		debugBody = sanitizeDebugBody(b)
	}
	req, err := http.NewRequest(method, u, r)
	if err != nil {
		return nil, err
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("X-SCENE", c.cfg.Scene)
	req.Header.Set("X-LANG", apiLang(c.cfg.Lang))
	if auth && c.token != "" {
		req.Header.Set("X-Auth-Token", c.token)
	}
	for name, values := range headers {
		for _, value := range values {
			req.Header.Add(name, value)
		}
	}

	start := time.Now()
	httpResp, err := c.httpClient.Do(req)
	if err != nil {
		c.debug(method, u, debugBody, 0, time.Since(start), "", err)
		return nil, err
	}
	defer httpResp.Body.Close()
	respBody, err := io.ReadAll(httpResp.Body)
	if err != nil {
		c.debug(method, u, debugBody, httpResp.StatusCode, time.Since(start), httpResp.Header.Get("X-Request-Id"), err)
		return nil, err
	}
	if httpResp.StatusCode < 200 || httpResp.StatusCode > 299 {
		httpErr := HTTPError{StatusCode: httpResp.StatusCode, Message: errorMessage(httpResp, respBody)}
		c.debug(method, u, debugBody, httpResp.StatusCode, time.Since(start), "", httpErr)
		return nil, httpErr
	}
	c.debug(method, u, debugBody, httpResp.StatusCode, time.Since(start), "", nil)
	return respBody, nil
}

func (c *Client) doWithHeaders(method, path string, q url.Values, body interface{}, auth bool, headers http.Header) (APIResponse, error) {
	u, err := c.url(path, q)
	if err != nil {
		return APIResponse{}, err
	}
	var r io.Reader
	debugBody := ""
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return APIResponse{}, err
		}
		r = bytes.NewReader(b)
		debugBody = sanitizeDebugBody(b)
	}
	req, err := http.NewRequest(method, u, r)
	if err != nil {
		return APIResponse{}, err
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("X-SCENE", c.cfg.Scene)
	req.Header.Set("X-LANG", apiLang(c.cfg.Lang))
	if auth && c.token != "" {
		req.Header.Set("X-Auth-Token", c.token)
	}
	for name, values := range headers {
		for _, value := range values {
			req.Header.Add(name, value)
		}
	}

	start := time.Now()
	httpResp, err := c.httpClient.Do(req)
	if err != nil {
		c.debug(method, u, debugBody, 0, time.Since(start), "", err)
		return APIResponse{}, err
	}
	defer httpResp.Body.Close()
	respBody, err := io.ReadAll(httpResp.Body)
	if err != nil {
		c.debug(method, u, debugBody, httpResp.StatusCode, time.Since(start), httpResp.Header.Get("X-Request-Id"), err)
		return APIResponse{}, err
	}
	if httpResp.StatusCode < 200 || httpResp.StatusCode > 299 {
		httpErr := HTTPError{StatusCode: httpResp.StatusCode, Message: errorMessage(httpResp, respBody)}
		c.debug(method, u, debugBody, httpResp.StatusCode, time.Since(start), "", httpErr)
		return APIResponse{}, httpErr
	}
	apiResp, err := decodeAPIResponse(respBody)
	if err != nil {
		c.debug(method, u, debugBody, httpResp.StatusCode, time.Since(start), "", err)
		return APIResponse{}, err
	}
	if !isSuccessCode(apiResp.Code) {
		apiErr := APIError{Code: apiResp.Code, Message: apiErrorMessage(apiResp)}
		c.debug(method, u, debugBody, httpResp.StatusCode, time.Since(start), apiResp.TraceID, apiErr)
		return APIResponse{}, apiErr
	}
	c.debug(method, u, debugBody, httpResp.StatusCode, time.Since(start), apiResp.TraceID, nil)
	return apiResp, nil
}

func apiLang(lang string) string {
	if lang == "zh_cn" {
		return "zh"
	}
	return lang
}

func (c *Client) debug(method, requestURL, body string, status int, duration time.Duration, traceID string, err error) {
	if !c.cfg.Debug || c.debugOut == nil {
		return
	}
	parts := []string{
		"DEBUG",
		"method=" + method,
		"url=" + requestURL,
		fmt.Sprintf("status=%d", status),
		"duration=" + duration.Round(time.Millisecond).String(),
	}
	if body != "" {
		parts = append(parts, "body="+body)
	}
	if traceID != "" {
		parts = append(parts, "trace_id="+traceID)
	}
	if err != nil {
		parts = append(parts, "error="+err.Error())
	}
	fmt.Fprintln(c.debugOut, strings.Join(parts, " "))
}

func sanitizeDebugBody(body []byte) string {
	var raw interface{}
	if err := json.Unmarshal(body, &raw); err != nil {
		return string(body)
	}
	sanitized := sanitizeDebugValue(raw)
	b, err := json.Marshal(sanitized)
	if err != nil {
		return string(body)
	}
	return string(b)
}

func sanitizeDebugValue(v interface{}) interface{} {
	switch t := v.(type) {
	case map[string]interface{}:
		out := make(map[string]interface{}, len(t))
		for key, value := range t {
			if sensitiveDebugKeys[key] {
				out[key] = "***"
				continue
			}
			out[key] = sanitizeDebugValue(value)
		}
		return out
	case []interface{}:
		out := make([]interface{}, len(t))
		for i, value := range t {
			out[i] = sanitizeDebugValue(value)
		}
		return out
	default:
		return v
	}
}

func (c *Client) url(path string, q url.Values) (string, error) {
	if c.cfg.Host == "" {
		return "", fmt.Errorf("host is required")
	}
	base, err := url.Parse(c.cfg.Host)
	if err != nil {
		return "", err
	}
	pathQuery := url.Values{}
	if idx := strings.Index(path, "?"); idx >= 0 {
		rawQuery := path[idx+1:]
		path = path[:idx]
		if rawQuery != "" {
			pathQuery, err = url.ParseQuery(rawQuery)
			if err != nil {
				return "", err
			}
		}
	}
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	base.Path = strings.TrimRight(base.Path, "/") + path
	mergedQuery := url.Values{}
	for key, values := range pathQuery {
		for _, value := range values {
			mergedQuery.Add(key, value)
		}
	}
	for key, values := range q {
		for _, value := range values {
			mergedQuery.Add(key, value)
		}
	}
	base.RawQuery = mergedQuery.Encode()
	return base.String(), nil
}

func decodeAPIResponse(b []byte) (APIResponse, error) {
	if len(bytes.TrimSpace(b)) == 0 {
		return APIResponse{}, nil
	}
	var decoded interface{}
	if err := json.Unmarshal(b, &decoded); err != nil {
		return APIResponse{}, err
	}
	raw, ok := decoded.(map[string]interface{})
	if !ok {
		return APIResponse{Data: decoded}, nil
	}
	return APIResponse{
		Code:    stringValue(raw["code"]),
		Data:    raw["data"],
		Error:   raw["error"],
		Title:   stringValue(raw["title"]),
		TraceID: stringValue(raw["trace_id"]),
		Raw:     raw,
	}, nil
}

func errorMessage(resp *http.Response, body []byte) string {
	if v := resp.Header.Get("Server-Error-Message"); v != "" {
		return v
	}
	var raw map[string]interface{}
	if err := json.Unmarshal(body, &raw); err == nil {
		if v := stringValue(raw["faultstring"]); v != "" {
			return v
		}
	}
	return string(bytes.TrimSpace(body))
}

func apiErrorMessage(resp APIResponse) string {
	if resp.Title != "" {
		return resp.Title
	}
	if s := stringValue(resp.Error); s != "" {
		return s
	}
	if resp.Error != nil {
		b, err := json.Marshal(resp.Error)
		if err == nil && string(b) != "{}" {
			return string(b)
		}
	}
	return ""
}

func isSuccessCode(code string) bool {
	switch code {
	case "", "00000000", "200":
		return true
	default:
		return false
	}
}

func isUnauthorized(err error) bool {
	if e, ok := err.(HTTPError); ok {
		return e.StatusCode == http.StatusUnauthorized
	}
	return false
}

func stringValue(v interface{}) string {
	if s, ok := v.(string); ok {
		return s
	}
	return ""
}
