package nws

import (
	"net/http"
	"time"
)

func defaultTransport() *http.Transport {
	t := http.DefaultTransport.(*http.Transport).Clone()
	t.MaxIdleConns = 100
	t.MaxConnsPerHost = 100
	t.MaxIdleConnsPerHost = 100
	return t
}

func defaultHTTP() *http.Client {
	return &http.Client{
		Transport: defaultTransport(),
		Timeout:   30 * time.Second,
	}
}
