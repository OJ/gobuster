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

// HTTPClient represents a http object
type HTTPClient struct {
	client        *http.Client
	context       context.Context
	userAgent     string
	username      string
	password      string
	includeLength bool
}

// HTTPOptions provides options to the http client
type HTTPOptions struct {
	Proxy          string
	Username       string
	Password       string
	UserAgent      string
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
			Proxy: proxyURLFunc,
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: opt.InsecureSSL,
			},
		}}
	client.context = c
	client.username = opt.Username
	client.password = opt.Password
	client.includeLength = opt.IncludeLength
	client.userAgent = opt.UserAgent
	return &client, nil
}

// Get makes an http request and returns the status, the length and an error
func (client *HTTPClient) Get(fullURL, host, cookie string) (*int, *int64, error) {
	resp, err := client.makeRequest(fullURL, host, cookie)
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

// GetBody makes an http request and returns the status and the body
func (client *HTTPClient) GetBody(fullURL, host, cookie string) (*int, *string, error) {
	resp, err := client.makeRequest(fullURL, host, cookie)
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
	bodyString := string(body)

	return &resp.StatusCode, &bodyString, nil
}

func (client *HTTPClient) makeRequest(fullURL, host, cookie string) (*http.Response, error) {
	req, err := http.NewRequest(http.MethodGet, fullURL, nil)
	if err != nil {
		return nil, err
	}

	// add the context so we can easily cancel out
	req = req.WithContext(client.context)

	if cookie != "" {
		req.Header.Set("Cookie", cookie)
	}

	if host != "" {
		req.Host = host
	}

	ua := DefaultUserAgent()
	if client.userAgent != "" {
		ua = client.userAgent
	}
	req.Header.Set("User-Agent", ua)

	if client.username != "" {
		req.SetBasicAuth(client.username, client.password)
	}

	resp, err := client.client.Do(req)
	if err != nil {
		if ue, ok := err.(*url.Error); ok {
			if strings.HasPrefix(ue.Err.Error(), "x509") {
				return nil, fmt.Errorf("Invalid certificate: %v", ue.Err)
			}
		}
		return nil, err
	}
	return resp, nil
}
