package client

import (
	"net/http"
	"time"
)

type HttpClient interface {
	GetBaseURL() string
	GetDefaultRequestTimeoutMS() time.Duration
	GetHttpClient() *http.Client
}
