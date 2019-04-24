package checkpoints

import (
	"net/http"
	"testing"
)

func TestResponseHeaderCheckpoint_ResponseValue(t *testing.T) {
	resp := new(http.Response)
	resp.StatusCode = 200
	resp.Header = http.Header{}
	resp.Header.Set("Hello", "World")

	checkpoint := new(ResponseHeaderCheckpoint)
	t.Log(checkpoint.ResponseValue(nil, resp, "Hello"))
}
