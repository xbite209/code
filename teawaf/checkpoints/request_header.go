package checkpoints

import (
	"github.com/TeaWeb/code/teawaf/requests"
	"net/http"
	"strings"
)

type RequestHeaderCheckpoint struct {
	Checkpoint
}

func (this *RequestHeaderCheckpoint) RequestValue(req *requests.Request, param string) (value interface{}, sysErr error, userErr error) {
	v, found := req.Header[param]
	if !found {
		value = ""
		return
	}
	value = strings.Join(v, ";")
	return
}

func (this *RequestHeaderCheckpoint) ResponseValue(req *requests.Request, resp *http.Response, param string) (value interface{}, sysErr error, userErr error) {
	if this.IsRequest() {
		return this.RequestValue(req, param)
	}
	return
}
