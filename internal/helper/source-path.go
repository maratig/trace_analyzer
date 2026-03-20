package helper

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"

	apiError "github.com/maratig/trace_analyzer/api/error"
	"golang.org/x/exp/trace"
)

const defaultHttpListeningSeconds = 36000

func CreateTraceReader(
	ctx context.Context, sourcePath string, endpointConnectionWait time.Duration,
) (*trace.Reader, io.Closer, error) {
	if ctx == nil {
		return nil, nil, errors.New("ctx must not be nil")
	}
	if sourcePath == "" {
		return nil, nil, errors.New("sourcePath must not be empty")
	}
	if endpointConnectionWait <= 0 {
		return nil, nil, errors.New("endpointConnectionWait must be greater than zero")
	}

	// Check if sourcePath is a valid url
	u, err := url.Parse(sourcePath)
	if err != nil {
		return nil, nil, fmt.Errorf("sourcePath is not a valid URL, sourcePath=%s; %w", sourcePath, err)
	}
	if u.Host == "" {
		return nil, nil, fmt.Errorf("host must be present in the given sourcePath=%s", sourcePath)
	}

	r, closer, err := createHttpReader(u)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create an http reader; %w", err)
	}

	return r, closer, nil
}

func createHttpReader(u *url.URL) (*trace.Reader, io.Closer, error) {
	params := u.Query()
	params.Set("seconds", strconv.Itoa(defaultHttpListeningSeconds))
	u.RawQuery = params.Encode()
	urlStr := u.String()

	resp, err := http.Get(urlStr)
	if err != nil {
		if _, ok := errors.AsType[*url.Error](err); ok {
			return nil, nil, fmt.Errorf("%w; %v", apiError.ErrConnectionFailed, err)
		}
		return nil, nil, fmt.Errorf("failed to get response from the given url; %w", err)
	}
	if resp.StatusCode >= 500 && resp.StatusCode < 600 {
		data, err := io.ReadAll(resp.Body)
		if err != nil {
			err = fmt.Errorf("failed to read response body; %w", err)
		}
		resp.Body.Close()
		return nil, nil, fmt.Errorf(
			"server error, status code=%d. body=%s; %w; %v",
			resp.StatusCode, string(data), apiError.ErrConnectionFailed, err,
		)
	}

	r := bufio.NewReader(resp.Body)
	ret, err := trace.NewReader(r)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create trace reader from url sourcePath; %w", err)
	}

	return ret, resp.Body, nil
}
