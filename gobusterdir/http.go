package gobusterdir

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"unicode/utf8"

	"github.com/OJ/gobuster/v3/libgobuster"
)

type httpClient struct {
	client        *http.Client
	context       context.Context
	userAgent     string
	username      string
	password      string
	includeLength bool
}

// NewHTTPClient returns a new HTTPClient
func newHTTPClient(c context.Context, opt *OptionsDir) (*httpClient, error) {
	var proxyURLFunc func(*http.Request) (*url.URL, error)
	var client httpClient
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

// MakeRequest makes a request to the specified url
func (client *httpClient) makeRequest(fullURL, cookie string, headers []string) (*int, *int64, error) {
	req, err := http.NewRequest(http.MethodGet, fullURL, nil)

	if err != nil {
		return nil, nil, err
	}

	// add the context so we can easily cancel out
	req = req.WithContext(client.context)

	if cookie != "" {
		req.Header.Set("Cookie", cookie)
	}

	ua := libgobuster.DefaultUserAgent()
	if client.userAgent != "" {
		ua = client.userAgent
	}
	req.Header.Set("User-Agent", ua)

	if len(headers) > 0 {
		for _,r := range headers {

			//parse Header
			keyAndValue := strings.SplitN(r,":",2)
			if len(keyAndValue)!=2 { //Check if we have both elements, header name and header value
				return nil, nil, fmt.Errorf("Error when parsing HTTP Headers: %v", r)
			}

			if len(keyAndValue[0])==0{ //Check that the header name is not empty, btw header value empty is ok
				return nil, nil, fmt.Errorf("Error when parsing HTTP Headers, header name is empty %v", r)
			}

			//req.Header.Set(keyAndValue[0], keyAndValue[1])
			//This is because Header.Set is case insensitive, and it always converts case by default, hEAdEr -> Header
			//So, I think this Idea is better:
			req.Header[keyAndValue[0]] = []string{keyAndValue[1]}


		}
	}

	if client.username != "" {
		req.SetBasicAuth(client.username, client.password)
	}

	resp, err := client.client.Do(req)
	if err != nil {
		if ue, ok := err.(*url.Error); ok {

			if strings.HasPrefix(ue.Err.Error(), "x509") {
				return nil, nil, fmt.Errorf("Invalid certificate: %v", ue.Err)
			}
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
		_, err = io.Copy(ioutil.Discard, resp.Body)
		if err != nil {
			return nil, nil, err
		}
	}

	return &resp.StatusCode, length, nil
}
