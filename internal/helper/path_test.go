package helper

import (
	"net/url"
	"testing"
)

func TestPath(t *testing.T) {
	p := "/some/path"
	u, err := url.Parse(p)
	if err != nil {
		t.Fatal(err)
	}
	println(u.Host)
}
