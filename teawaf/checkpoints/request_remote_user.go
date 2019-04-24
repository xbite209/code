package checkpoints

import (
	"github.com/TeaWeb/code/teawaf/requests"
	"net/http"
)

type RequestRemoteUserCheckpoint struct {
	Checkpoint
}

func (this *RequestRemoteUserCheckpoint) RequestValue(req *requests.Request, param string) (value interface{}, sysErr error, userErr error) {
	username, _, ok := req.BasicAuth()
	if !ok {
		value = ""
		return
	}
	value = username
	return
}

func (this *RequestRemoteUserCheckpoint) ResponseValue(req *requests.Request, resp *http.Response, param string) (value interface{}, sysErr error, userErr error) {
	if this.IsRequest() {
		return this.RequestValue(req, param)
	}
	return
}
