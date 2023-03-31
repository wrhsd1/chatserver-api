/*
 * @Author: cloudyi.li
 * @Date: 2023-03-30 18:16:23
 * @LastEditTime: 2023-03-31 17:04:24
 * @LastEditors: cloudyi.li
 * @FilePath: /chatserver-api/pkg/openai/chat_stream.go
 */
package openai

import (
	"bufio"
)

type ChatCompletionStreamChoiceDelta struct {
	Content string `json:"content"`
}

type ChatCompletionStreamChoice struct {
	Index        int                             `json:"index"`
	Delta        ChatCompletionStreamChoiceDelta `json:"delta"`
	FinishReason string                          `json:"finish_reason"`
}

type ChatCompletionStreamResponse struct {
	ID      string                       `json:"id"`
	Object  string                       `json:"object"`
	Created int64                        `json:"created"`
	Model   string                       `json:"model"`
	Choices []ChatCompletionStreamChoice `json:"choices"`
}

// ChatCompletionStream
// Note: Perhaps it is more elegant to abstract Stream using generics.
type ChatCompletionStream struct {
	*streamReader[ChatCompletionStreamResponse]
}

// CreateChatCompletionStream — API call to create a chat completion w/ streaming
// support. It sets whether to stream back partial progress. If set, tokens will be
// sent as data-only server-sent events as they become available, with the
// stream terminated by a data: [DONE] message.
func (c *Client) CreateChatCompletionStream(
	request ChatCompletionRequest,
) (stream *ChatCompletionStream, err error) {
	urlSuffix := "/chat/completions"
	if !checkEndpointSupportsModel(urlSuffix, request.Model) {
		err = ErrChatCompletionInvalidModel
		return
	}

	request.Stream = true
	req, err := c.newStreamRequest("POST", urlSuffix, request)
	if err != nil {
		return
	}

	resp, err := c.config.HTTPClient.Do(req) //nolint:bodyclose // body is closed in stream.Close()
	if err != nil {
		return
	}

	stream = &ChatCompletionStream{
		streamReader: &streamReader[ChatCompletionStreamResponse]{
			emptyMessagesLimit: c.config.EmptyMessagesLimit,
			reader:             bufio.NewReader(resp.Body),
			response:           resp,
			errAccumulator:     newErrorAccumulator(),
			unmarshaler:        &jsonUnmarshaler{},
		},
	}
	return
}