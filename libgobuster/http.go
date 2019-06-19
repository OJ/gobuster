package libgobuster

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"
	"unicode/utf8"
)

// HTTPHeader holds a single key value pair of a HTTP header
type HTTPHeader struct {
	Name  string
	Value string
}

// HTTPClient represents a http object
type HTTPClient struct {
	client           *http.Client
	context          context.Context
	userAgent        string
	defaultUserAgent string
	username         string
	password         string
	headers          []HTTPHeader
	includeLength    bool
}

// HTTPOptions provides options to the http client
type HTTPOptions struct {
	Proxy          string
	Username       string
	Password       string
	UserAgent      string
	Headers        []HTTPHeader
	Timeout        time.Duration
	FollowRedirect bool
	InsecureSSL    bool
	IncludeLength  bool
}

// NewHTTPClient returns a new HTTPClient
func NewHTTPClient(c context.Context, opt *HTTPOptions) (*HTTPClient, error) {
	var proxyURLFunc func(*http.Request) (*url.URL, error)
	var client HTTPClient
	proxyURLFunc = http.ProxyFromEnvironment

	if opt == nil {
		return nil, fmt.Errorf("options is nil")
	}

	if opt.Proxy != "" {
		proxyURL, err := url.Parse(opt.Proxy)
		if err != nil {
			return nil, fmt.Errorf("proxy URL is invalid (%v)", err)
		}
		proxyURLFunc = http.ProxyURL(proxyURL)
	}

	var redirectFunc func(req *http.Request, via []*http.Request) error
	if !opt.FollowRedirect {
		redirectFunc = func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		}
	} else {
		redirectFunc = nil
	}

	client.client = &http.Client{
		Timeout:       opt.Timeout,
		CheckRedirect: redirectFunc,
		Transport: &http.Transport{
			Proxy:               proxyURLFunc,
			MaxIdleConns:        100,
			MaxIdleConnsPerHost: 100,
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: opt.InsecureSSL,
			},
		}}
	client.context = c
	client.username = opt.Username
	client.password = opt.Password
	client.includeLength = opt.IncludeLength
	client.userAgent = opt.UserAgent
	client.defaultUserAgent = DefaultUserAgent()
	client.headers = opt.Headers
	return &client, nil
}

// Get gets an URL and returns the status, the length and an error
func (client *HTTPClient) Get(fullURL, host, cookie string) (*int, *int64, error) {
	return client.requestWithoutBody(http.MethodGet, fullURL, host, cookie, nil)
}

// Post posts to an URL and returns the status, the length and an error
func (client *HTTPClient) Post(fullURL, host, cookie string, data io.Reader) (*int, *int64, error) {
	return client.requestWithoutBody(http.MethodPost, fullURL, host, cookie, data)
}

// GetWithBody gets an URL and returns the status and the body
func (client *HTTPClient) GetWithBody(fullURL, host, cookie string) (*int, *[]byte, error) {
	return client.requestWithBody(http.MethodGet, fullURL, host, cookie, nil)
}

// PostWithBody gets an URL and returns the status and the body
func (client *HTTPClient) PostWithBody(fullURL, host, cookie string, data io.Reader) (*int, *[]byte, error) {
	return client.requestWithBody(http.MethodPost, fullURL, host, cookie, data)
}

// requestWithoutBody makes an http request and returns the status, the length and an error
func (client *HTTPClient) requestWithoutBody(method, fullURL, host, cookie string, data io.Reader) (*int, *int64, error) {
	resp, err := client.makeRequest(method, fullURL, host, cookie, data)
	if err != nil {
		// ignore context canceled errors
		if client.context.Err() == context.Canceled {
			return nil, nil, nil
		}
		return nil, nil, err
	}
	defer resp.Body.Close()

	var length *int64

	if client.includeLength {
		length = new(int64)
		if resp.ContentLength <= 0 {
			body, err2 := ioutil.ReadAll(resp.Body)
			if err2 == nil {
				*length = int64(utf8.RuneCountInString(string(body)))
			}
		} else {
			*length = resp.ContentLength
		}
	} else {
		// DO NOT REMOVE!
		// absolutely needed so golang will reuse connections!
		_, err := io.Copy(ioutil.Discard, resp.Body)
		if err != nil {
			return nil, nil, err
		}
	}

	return &resp.StatusCode, length, nil
}

// requestWithBody makes an http request and returns the status and the body
func (client *HTTPClient) requestWithBody(method, fullURL, host, cookie string, data io.Reader) (*int, *[]byte, error) {
	resp, err := client.makeRequest(method, fullURL, host, cookie, data)
	if err != nil {
		// ignore context canceled errors
		if client.context.Err() == context.Canceled {
			return nil, nil, nil
		}
		return nil, nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, nil, fmt.Errorf("could not read body: %v", err)
	}

	return &resp.StatusCode, &body, nil
}

func (client *HTTPClient) makeRequest(method, fullURL, host, cookie string, data io.Reader) (*http.Response, error) {
	var req *http.Request
	var err error

	switch method {
	case http.MethodGet:
		req, err = http.NewRequest(http.MethodGet, fullURL, nil)
		if err != nil {
			return nil, err
		}
	case http.MethodPost:
		req, err = http.NewRequest(http.MethodPost, fullURL, data)
		if err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("invalid method %s", method)
	}

	// add the context so we can easily cancel out
	req = req.WithContext(client.context)

	if cookie != "" {
		req.Header.Set("Cookie", cookie)
	}

	if host != "" {
		req.Host = host
	}

	if client.userAgent != "" {
		req.Header.Set("User-Agent", client.userAgent)
	} else {
		req.Header.Set("User-Agent", client.defaultUserAgent)
	}

	// add custom headers
	for _, h := range client.headers {
		req.Header.Set(h.Name, h.Value)
	}

	if client.username != "" {
		req.SetBasicAuth(client.username, client.password)
	}

	resp, err := client.client.Do(req)
	if err != nil {
		if ue, ok := err.(*url.Error); ok {
			if strings.HasPrefix(ue.Err.Error(), "x509") {
				return nil, fmt.Errorf("invalid certificate: %v", ue.Err)
			}
		}
		return nil, err
	}

	return resp, nil
}
