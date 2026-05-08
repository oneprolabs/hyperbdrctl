package commands

import (
	"net/url"

	"hyperbdr-client/internal/client"
)

type commandAPIAdapter struct {
	ctx *context
}

type commandPosterAdapter struct {
	ctx *context
}

func (a commandAPIAdapter) Get(path string, q url.Values) (client.APIResponse, error) {
	return fetch(a.ctx, path, q)
}

func (a commandAPIAdapter) Post(path string, body interface{}) (client.APIResponse, error) {
	return post(a.ctx, path, body)
}

func (a commandAPIAdapter) Delete(path string, body interface{}) (client.APIResponse, error) {
	return deleteWithBody(a.ctx, path, body)
}

func (a commandPosterAdapter) Post(path string, body interface{}) (client.APIResponse, error) {
	return post(a.ctx, path, body)
}

func (a commandPosterAdapter) Get(path string, q url.Values) (client.APIResponse, error) {
	return fetch(a.ctx, path, q)
}

func (a commandPosterAdapter) Delete(path string, body interface{}) (client.APIResponse, error) {
	return deleteWithBody(a.ctx, path, body)
}
