package checkpoints

import (
	"net/http"
	"testing"
)

func TestResponseStatusCheckpoint_ResponseValue(t *testing.T) {
	resp := new(http.Response)
	resp.StatusCode = 200

	checkpoint := new(ResponseStatusCheckpoint)
	t.Log(checkpoint.ResponseValue(nil, resp, ""))
}
