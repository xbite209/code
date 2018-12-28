package backend

import (
	"github.com/TeaWeb/code/teaconfigs"
	"github.com/TeaWeb/code/teaconfigs/scheduling"
	"github.com/iwind/TeaGo/actions"
)

type IndexAction actions.Action

// 后端列表
func (this *IndexAction) Run(params struct {
	Server string
}) {
	server, err := teaconfigs.NewServerConfigFromFile(params.Server)
	if err != nil {
		this.Fail(err.Error())
	}

	this.Data["selectedTab"] = "backend"
	this.Data["filename"] = params.Server
	this.Data["proxy"] = server

	normalBackends := []*teaconfigs.ServerBackendConfig{}
	backupBackends := []*teaconfigs.ServerBackendConfig{}
	for _, backend := range server.Backends {
		if backend.IsBackup {
			backupBackends = append(backupBackends, backend)
		} else {
			normalBackends = append(normalBackends, backend)
		}
	}

	this.Data["normalBackends"] = normalBackends
	this.Data["backupBackends"] = backupBackends

	// 算法
	if server.Scheduling == nil {
		this.Data["scheduling"] = scheduling.FindSchedulingType("random")
	} else {
		s := scheduling.FindSchedulingType(server.Scheduling.Code)
		if s == nil {
			this.Data["scheduling"] = scheduling.FindSchedulingType("random")
		} else {
			this.Data["scheduling"] = s
		}
	}

	this.Show()
}
