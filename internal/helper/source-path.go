package helper

import (
	"bufio"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"

	"golang.org/x/exp/trace"
)

func CreateTraceReader(sourcePath string) (*trace.Reader, io.Closer, error) {
	// Check if sourcePath is an url
	u, err := url.Parse(sourcePath)
	if err == nil && u.Host != "" {
		// TODO process http codes and errors more accurately
		resp, err := http.Get(sourcePath)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to get response from sourcePath; %w", err)
		}
		r := bufio.NewReader(resp.Body)
		ret, err := trace.NewReader(r)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to create trace reader from url sourcePath; %w", err)
		}
		return ret, resp.Body, nil
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
