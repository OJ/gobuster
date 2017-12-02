// Rediret handler

package main

import "net/http"

type redirectHandler struct {
	Transport http.RoundTripper
	Config    *config
}

func (rh *redirectHandler) RoundTrip(req *http.Request) (resp *http.Response, err error) {
	if rh.Config.FollowRedirect {
		return rh.Transport.RoundTrip(req)
	}

	resp, err = rh.Transport.RoundTrip(req)
	if err != nil {
		return resp, err
	}

	switch resp.StatusCode {
	case http.StatusMovedPermanently, http.StatusFound, http.StatusSeeOther,
		http.StatusNotModified, http.StatusUseProxy, http.StatusTemporaryRedirect:
		return nil, &redirectError{StatusCode: resp.StatusCode}
	}

	return resp, err
}
