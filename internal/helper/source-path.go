package helper

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"time"

	"golang.org/x/exp/trace"
)

const defaultHttpListeningSeconds = 36000

func CreateTraceReader(
	ctx context.Context, sourcePath string, endpointConnectionWaitSec int,
) (*trace.Reader, io.Closer, error) {
	if ctx == nil {
		return nil, nil, errors.New("ctx must not be nil")
	}
	if sourcePath == "" {
		return nil, nil, errors.New("sourcePath must not be empty")
	}
	if endpointConnectionWaitSec <= 0 {
		return nil, nil, errors.New("endpointConnectionWaitSec must be greater than zero")
	}

	// Check if sourcePath is a url
	u, err := url.Parse(sourcePath)
	if err == nil && u.Host != "" {
		r, closer, err := createHttpReader(ctx, u, endpointConnectionWaitSec)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to create an http reader; %w", err)
		}
		return r, closer, nil
	}

	// Check if sourcePath is a local path
	f, err := os.Open(sourcePath)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to open sourcePath as a file; %w", err)
	}
	r := bufio.NewReader(f)
	ret, err := trace.NewReader(r)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create trace reader from file sourcePath; %w", err)
	}

	return ret, f, nil
}

func createHttpReader(
	ctx context.Context, u *url.URL, endpointConnectionWaitSec int,
) (*trace.Reader, io.Closer, error) {
	localCtx, cancel := context.WithTimeout(ctx, time.Duration(endpointConnectionWaitSec)*time.Second)
	defer cancel()

	params := u.Query()
	params.Set("seconds", strconv.Itoa(defaultHttpListeningSeconds))
	u.RawQuery = params.Encode()
	urlStr := u.String()
	for {
		if localCtx.Err() != nil {
			return nil, nil, localCtx.Err()
		}

		resp, err := http.Get(urlStr)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to get response from the given url; %w", err)
		}
		if resp.StatusCode >= 500 && resp.StatusCode < 600 {
			time.Sleep(5 * time.Millisecond)
			continue
		}

		r := bufio.NewReader(resp.Body)
		ret, err := trace.NewReader(r)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to create trace reader from url sourcePath; %w", err)
		}
		return ret, resp.Body, nil
	}
}
