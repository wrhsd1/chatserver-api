/*
 * @Author: cloudyi.li
 * @Date: 2023-03-30 18:16:24
 * @LastEditTime: 2023-05-12 23:21:09
 * @LastEditors: cloudyi.li
 * @FilePath: /chatserver-api/pkg/openai/edits.go
 */
package openai

import (
	"fmt"
	"net/http"
)

// EditsRequest represents a request structure for Edits API.
type EditsRequest struct {
	Model       *string `json:"model,omitempty"`
	Input       string  `json:"input,omitempty"`
	Instruction string  `json:"instruction,omitempty"`
	N           int     `json:"n,omitempty"`
	Temperature float32 `json:"temperature,omitempty"`
	TopP        float32 `json:"top_p,omitempty"`
}

// EditsChoice represents one of possible edits.
type EditsChoice struct {
	Text  string `json:"text"`
	Index int    `json:"index"`
}

// EditsResponse represents a response structure for Edits API.
type EditsResponse struct {
	Object  string        `json:"object"`
	Created int64         `json:"created"`
	Usage   Usage         `json:"usage"`
	Choices []EditsChoice `json:"choices"`
}

// Perform an API call to the Edits endpoint.
func (c *Client) Edits(request EditsRequest) (response EditsResponse, err error) {
	req, err := c.requestBuilder.build(c.ctx, http.MethodPost, c.fullURL("/edits", fmt.Sprint(request.Model)), request)
	if err != nil {
		return
	}

	err = c.sendRequest(req, &response)
	return
}
