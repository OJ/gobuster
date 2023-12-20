package libgobuster

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"
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
	client                *http.Client
	userAgent             string
	defaultUserAgent      string
	username              string
	password              string
	headers               []HTTPHeader
	noCanonicalizeHeaders bool
	cookies               string
	method                string
	host                  string
	logger                *Logger
}

// RequestOptions is used to pass options to a single individual request
type RequestOptions struct {
	Host                     string
	Body                     io.Reader
	ReturnBody               bool
	ModifiedHeaders          []HTTPHeader
	UpdatedBasicAuthUsername string
	UpdatedBasicAuthPassword string
}

// NewHTTPClient returns a new HTTPClient
func NewHTTPClient(opt *HTTPOptions, logger *Logger) (*HTTPClient, error) {
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

	tlsConfig := tls.Config{
		InsecureSkipVerify: opt.NoTLSValidation,
		// enable TLS1.0 and TLS1.1 support
		MinVersion: tls.VersionTLS10,
	}
	if opt.TLSCertificate != nil {
		tlsConfig.Certificates = []tls.Certificate{*opt.TLSCertificate}
	}

	client.client = &http.Client{
		Timeout:       opt.Timeout,
		CheckRedirect: redirectFunc,
		Transport: &http.Transport{
			Proxy:               proxyURLFunc,
			MaxIdleConns:        100,
			MaxIdleConnsPerHost: 100,
			TLSClientConfig:     &tlsConfig,
		}}
	client.username = opt.Username
	client.password = opt.Password
	client.userAgent = opt.UserAgent
	client.defaultUserAgent = DefaultUserAgent()
	client.headers = opt.Headers
	client.noCanonicalizeHeaders = opt.NoCanonicalizeHeaders
	client.cookies = opt.Cookies
	client.method = opt.Method
	if client.method == "" {
		client.method = http.MethodGet
	}
	// Host header needs to be set separately
	for _, h := range opt.Headers {
		if h.Name == "Host" {
			client.host = h.Value
			break
		}
	}
	client.logger = logger
	return &client, nil
}

// Request makes a http request and returns the status, the content length, the headers, the body and an error
// if you want the body returned set the corresponding property inside RequestOptions
func (client *HTTPClient) Request(ctx context.Context, fullURL string, opts RequestOptions) (int, int64, http.Header, []byte, error) {
	resp, err := client.makeRequest(ctx, fullURL, opts)
	if err != nil {
		// ignore context canceled errors
		if errors.Is(ctx.Err(), context.Canceled) {
			return 0, 0, nil, nil, nil
		}
		return 0, 0, nil, nil, err
	}
	defer resp.Body.Close()

	var body []byte
	var length int64
	if opts.ReturnBody {
		body, err = io.ReadAll(resp.Body)
		if err != nil {
			return 0, 0, nil, nil, fmt.Errorf("could not read body %w", err)
		}
		length = int64(len(body))
	} else {
		// DO NOT REMOVE!
		// absolutely needed so golang will reuse connections!
		length, err = io.Copy(io.Discard, resp.Body)
		if err != nil {
			return 0, 0, nil, nil, err
		}
	}

	return resp.StatusCode, length, resp.Header, body, nil
}

func (client *HTTPClient) makeRequest(ctx context.Context, fullURL string, opts RequestOptions) (*http.Response, error) {
	req, err := http.NewRequest(client.method, fullURL, opts.Body)
	if err != nil {
		return nil, err
	}

	// add the context so we can easily cancel out
	req = req.WithContext(ctx)

	if client.cookies != "" {
		req.Header.Set("Cookie", client.cookies)
	}

	// Use host for VHOST mode on a per-request basis, otherwise the one provided from headers
	if opts.Host != "" {
		req.Host = opts.Host
	} else if client.host != "" {
		req.Host = client.host
	}

	if client.userAgent != "" {
		req.Header.Set("User-Agent", client.userAgent)
	} else {
		req.Header.Set("User-Agent", client.defaultUserAgent)
	}

	// add custom headers
	// if ModifiedHeaders are supplied use those, otherwise use the original ones
	// currently only relevant on fuzzing
	if len(opts.ModifiedHeaders) > 0 {
		for _, h := range opts.ModifiedHeaders {
			// empty headers are not valid (happens when fuzzing the host header for example because the slice is initialized with the provided header length)
			if h.Name == "" {
				continue
			}

			if client.noCanonicalizeHeaders {
				// https://stackoverflow.com/questions/26351716/how-to-keep-key-case-sensitive-in-request-header-using-golang
				req.Header[h.Name] = []string{h.Value}
			} else {
				req.Header.Set(h.Name, h.Value)
			}
		}
	} else {
		for _, h := range client.headers {
			if client.noCanonicalizeHeaders {
				// https://stackoverflow.com/questions/26351716/how-to-keep-key-case-sensitive-in-request-header-using-golang
				req.Header[h.Name] = []string{h.Value}
			} else {
				req.Header.Set(h.Name, h.Value)
			}
		}
	}

	if opts.UpdatedBasicAuthUsername != "" {
		req.SetBasicAuth(opts.UpdatedBasicAuthUsername, opts.UpdatedBasicAuthPassword)
	} else if client.username != "" {
		req.SetBasicAuth(client.username, client.password)
	}

	if client.logger.debug {
		dump, err := httputil.DumpRequestOut(req, false)
		if err != nil {
			return nil, err
		}
		client.logger.Debugf("%s", dump)
	}

	resp, err := client.client.Do(req)
	if err != nil {
		var ue *url.Error
		if errors.As(err, &ue) {
			if strings.HasPrefix(ue.Err.Error(), "x509") {
				return nil, fmt.Errorf("invalid certificate: %w", ue.Err)
			}
		}
		return nil, err
	}

	return resp, nil
}
