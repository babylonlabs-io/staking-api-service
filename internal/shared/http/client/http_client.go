package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/babylonlabs-io/staking-api-service/internal/shared/observability/metrics"
	"github.com/babylonlabs-io/staking-api-service/internal/shared/types"
	"github.com/rs/zerolog/log"
)

var ALLOWED_METHODS = []string{"GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS"}

type HttpClientOptions struct {
	Timeout      int
	Path         string
	TemplatePath string // Metrics purpose
	Headers      map[string]string
}

func isAllowedMethod(method string) bool {
	for _, allowedMethod := range ALLOWED_METHODS {
		if method == allowedMethod {
			return true
		}
	}
	return false
}

func sendRequest[I any, R any](
	ctx context.Context, client HttpClient, method string, opts *HttpClientOptions, input *I,
) (*R, *types.Error) {
	if !isAllowedMethod(method) {
		return nil, types.NewInternalServiceError(fmt.Errorf("method %s is not allowed", method))
	}
	url := fmt.Sprintf("%s%s", client.GetBaseURL(), opts.Path)
	timeout := client.GetDefaultRequestTimeout()
	// If timeout is set, use it instead of the default
	if opts.Timeout != 0 {
		timeout = opts.Timeout
	}
	// Set a timeout for the request
	ctxWithTimeout, cancel := context.WithTimeout(ctx, time.Duration(timeout)*time.Millisecond)
	defer cancel()

	var req *http.Request
	var requestError error
	if input != nil && (method == http.MethodPost || method == http.MethodPut) {
		body, err := json.Marshal(input)
		if err != nil {
			return nil, types.NewErrorWithMsg(
				http.StatusInternalServerError,
				types.InternalServiceError,
				"failed to marshal request body",
			)
		}
		req, requestError = http.NewRequestWithContext(ctxWithTimeout, method, url, bytes.NewBuffer(body))
	} else {
		req, requestError = http.NewRequestWithContext(ctxWithTimeout, method, url, nil)
	}
	if requestError != nil {
		return nil, types.NewErrorWithMsg(
			http.StatusInternalServerError, types.InternalServiceError, requestError.Error(),
		)
	}
	// Set headers
	for key, value := range opts.Headers {
		req.Header.Set(key, value)
	}

	resp, err := client.GetHttpClient().Do(req)
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded || err.Error() == "context canceled" {
			return nil, types.NewErrorWithMsg(
				http.StatusRequestTimeout,
				types.RequestTimeout,
				fmt.Sprintf("request timeout after %d ms at %s", timeout, url),
			)
		}
		return nil, types.NewErrorWithMsg(
			http.StatusInternalServerError,
			types.InternalServiceError,
			fmt.Sprintf("failed to send request to %s", url),
		)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= http.StatusInternalServerError {
		return nil, types.NewErrorWithMsg(
			resp.StatusCode,
			types.InternalServiceError,
			fmt.Sprintf("internal server error when calling %s", url),
		)
	} else if resp.StatusCode >= http.StatusBadRequest {
		return nil, types.NewErrorWithMsg(
			resp.StatusCode,
			types.BadRequest,
			fmt.Sprintf("client error when calling %s", url),
		)
	}

	var output R
	if err := json.NewDecoder(resp.Body).Decode(&output); err != nil {
		return nil, types.NewErrorWithMsg(
			http.StatusInternalServerError,
			types.InternalServiceError,
			fmt.Sprintf("failed to decode response from %s", url),
		)
	}

	return &output, nil
}

func SendRequest[I any, R any](
	ctx context.Context, client HttpClient, method string, opts *HttpClientOptions, input *I,
) (*R, *types.Error) {
	timer := metrics.StartClientRequestDurationTimer(
		client.GetBaseURL(), method, opts.TemplatePath,
	)
	result, err := sendRequest[I, R](ctx, client, method, opts, input)
	if err != nil {
		log.Ctx(ctx).Error().Err(err).Msgf("failed to send request")
		timer(err.StatusCode)
		return nil, err
	}
	timer(http.StatusOK)
	return result, err
}
