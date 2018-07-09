package libgobuster

import (
	"context"
	"testing"

	"github.com/h2non/gock"
)

func TestMakeRequest(t *testing.T) {
	defer gock.Off()
	gock.New("http://server.com").
		Get("/bar").
		Reply(200).
		BodyString("test")

	o := NewOptions()
	c, err := newHTTPClient(context.Background(), o)
	if err != nil {
		t.Fatalf("Got Error: %v", err)
	}
	gock.InterceptClient(c.client)
	defer gock.RestoreClient(c.client)
	a, b, err := c.makeRequest("http://server.com/bar", "")
	if err != nil {
		t.Fatalf("Got Error: %v", err)
	}
	if *a != 200 {
		t.Fatalf("Invalid status returned: %d", a)
	}
	if b != nil && *b != int64(len("test")) {
		t.Fatalf("Invalid length returned: %d", b)
	}
}
