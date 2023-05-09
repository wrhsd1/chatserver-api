package openai

import (
	"chatserver-api/pkg/config"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"golang.org/x/net/proxy"
)

// Client is OpenAI GPT-3 API client.
type Client struct {
	config            ClientConfig
	ctx               context.Context
	requestBuilder    requestBuilder
	createFormBuilder func(io.Writer) formBuilder
}

// NewClient creates new OpenAI API client.
func NewClient() (*Client, error) {
	c := DefaultConfig()
	config := config.AppConfig.OpenAIConfig
	if config.ProxyMode == "socks5" {
		proxyadd := fmt.Sprintf("%s:%s", config.ProxyIP, config.ProxyPort)
		dialer, err := proxy.SOCKS5("tcp", proxyadd, nil, proxy.Direct)
		if err != nil {
			return nil, err
		}
		transport := &http.Transport{
			Dial: dialer.Dial,
		}
		c.HTTPClient = &http.Client{
			Transport: transport,
		}

	}
	if config.ProxyMode == "http" {
		proxyUrl, _ := url.Parse(fmt.Sprintf("http://%s:%s", config.ProxyIP, config.ProxyPort))
		transport := &http.Transport{
			Proxy: http.ProxyURL(proxyUrl),
		}
		c.HTTPClient = &http.Client{
			Transport: transport,
		}
	}
	return NewClientWithConfig(c), nil
}

// NewClientWithConfig creates new OpenAI API client for specified config.
func NewClientWithConfig(config ClientConfig) *Client {
	return &Client{
		config:         config,
		ctx:            context.Background(),
		requestBuilder: newRequestBuilder(),
		createFormBuilder: func(body io.Writer) formBuilder {
			return newFormBuilder(body)
		},
	}
}

// NewOrgClient creates new OpenAI API client for specified Organization ID.
//
// Deprecated: Please use NewClientWithConfig.
func NewOrgClient() *Client {
	config := DefaultConfig()
	return NewClientWithConfig(config)
}

func (c *Client) sendRequest(req *http.Request, v interface{}) error {
	req.Header.Set("Accept", "application/json; charset=utf-8")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.config.authToken))

	// Check whether Content-Type is already set, Upload Files API requires
	// Content-Type == multipart/form-data
	contentType := req.Header.Get("Content-Type")
	if contentType == "" {
		req.Header.Set("Content-Type", "application/json; charset=utf-8")
	}

	if len(c.config.OrgID) > 0 {
		req.Header.Set("OpenAI-Organization", c.config.OrgID)
	}

	res, err := c.config.HTTPClient.Do(req)
	if err != nil {
		return err
	}

	defer res.Body.Close()

	if res.StatusCode < http.StatusOK || res.StatusCode >= http.StatusBadRequest {
		return c.handleErrorResp(res)
	}

	if v != nil {
		if err = json.NewDecoder(res.Body).Decode(v); err != nil {
			return err
		}
	}

	return nil
}

func (c *Client) fullURL(suffix string) string {
	return fmt.Sprintf("%s%s", c.config.BaseURL, suffix)
}

func (c *Client) newStreamRequest(
	method string,
	urlSuffix string,
	body any) (*http.Request, error) {
	req, err := c.requestBuilder.build(c.ctx, method, c.fullURL(urlSuffix), body)
	if err != nil {
		return nil, err
	}
	if len(c.config.OrgID) > 0 {
		req.Header.Set("OpenAI-Organization", c.config.OrgID)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "text/event-stream")
	req.Header.Set("Cache-Control", "no-cache")
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.config.authToken))

	return req, nil
}
func (c *Client) handleErrorResp(resp *http.Response) error {
	var errRes ErrorResponse
	err := json.NewDecoder(resp.Body).Decode(&errRes)
	if err != nil || errRes.Error == nil {
		reqErr := RequestError{
			HTTPStatusCode: resp.StatusCode,
			Err:            err,
		}
		return fmt.Errorf("error, %w", &reqErr)
	}
	errRes.Error.HTTPStatusCode = resp.StatusCode
	return fmt.Errorf("error, status code: %d, message: %w", resp.StatusCode, errRes.Error)
}
