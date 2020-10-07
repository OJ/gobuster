package libgobuster

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
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
	cookies          string
	method           string
}

// RequestOptions is used to pass options to a single individual request
type RequestOptions struct {
	Host       string
	Body       io.Reader
	ReturnBody bool
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
			return nil, fmt.Errorf("proxy URL is invalid (%w)", err)
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
				InsecureSkipVerify: opt.NoTLSValidation,
			},
		}}
	client.context = c
	client.username = opt.Username
	client.password = opt.Password
	client.userAgent = opt.UserAgent
	client.defaultUserAgent = DefaultUserAgent()
	client.headers = opt.Headers
	client.cookies = opt.Cookies
	client.method = opt.Method
	if client.method == "" {
		client.method = http.MethodGet
	}
	return &client, nil
}

// Request makes an http request and returns the status, the content length, the headers, the body and an error
// if you want the body returned set the corresponding property inside RequestOptions
func (client *HTTPClient) Request(fullURL string, opts RequestOptions) (*int, int64, http.Header, []byte, error) {
	resp, err := client.makeRequest(fullURL, opts.Host, opts.Body)
	if err != nil {
		// ignore context canceled errors
		if errors.Is(client.context.Err(), context.Canceled) {
			return nil, 0, nil, nil, nil
		}
		return nil, 0, nil, nil, err
	}
	defer resp.Body.Close()

	var body []byte
	var length int64
	if opts.ReturnBody {
		body, err = ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, 0, nil, nil, fmt.Errorf("could not read body %w", err)
		}
		length = int64(len(body))
	} else {
		// DO NOT REMOVE!
		// absolutely needed so golang will reuse connections!
		length, err = io.Copy(ioutil.Discard, resp.Body)
		if err != nil {
			return nil, 0, nil, nil, err
		}
	}

	return &resp.StatusCode, length, resp.Header, body, nil
}

func (client *HTTPClient) makeRequest(fullURL, host string, data io.Reader) (*http.Response, error) {
	req, err := http.NewRequest(client.method, fullURL, data)
	if err != nil {
		return nil, err
	}

	// add the context so we can easily cancel out
	req = req.WithContext(client.context)

	if client.cookies != "" {
		req.Header.Set("Cookie", client.cookies)
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
				return nil, fmt.Errorf("invalid certificate: %w", ue.Err)
			}
		}
		return nil, err
	}

	return resp, nil
}
