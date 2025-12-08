package client

import (
	"net/http"
	"time"
)

type HttpClient interface {
	GetBaseURL() string
	GetDefaultRequestTimeout() time.Duration
	GetHttpClient() *http.Client
}
