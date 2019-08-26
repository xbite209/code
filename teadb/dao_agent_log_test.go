package teadb

import (
	"github.com/TeaWeb/code/teaconfigs/agents"
	"github.com/iwind/TeaGo/assert"
	stringutil "github.com/iwind/TeaGo/utils/string"
	"testing"
	"time"
)

func TestAgentLogDAO_InsertOne(t *testing.T) {
	{
		log := new(agents.ProcessLog)
		log.AgentId = "test"
		log.Data = "abcdefg"
		log.TaskId = "abc"
		log.SetTime(time.Now())
		err := AgentLogDAO().InsertOne("test", log)
		if err != nil {
			t.Fatal(err)
		}
		t.Log("ok")
	}

	{
		log := new(agents.ProcessLog)
		log.AgentId = "test"
		log.Data = "abcdefg1"
		log.TaskId = "abc"
		log.SetTime(time.Now())
		err := AgentLogDAO().InsertOne("test", log)
		if err != nil {
			t.Fatal(err)
		}
		t.Log("ok")
	}

	{
		log := new(agents.ProcessLog)
		log.AgentId = "test"
		log.Data = "abcdefg2"
		log.TaskId = "abc"
		log.SetTime(time.Now())
		err := AgentLogDAO().InsertOne("test", log)
		if err != nil {
			t.Fatal(err)
		}
		t.Log("ok")
	}
}

func TestAgentLogDAO_ListTaskLogs(t *testing.T) {
	a := assert.NewAssertion(t)

	taskId := "abc"
	taskLogs, err := AgentLogDAO().ListTaskLogs("test", taskId, "", 2)
	if err != nil {
		t.Fatal(err)
	}

	for _, log := range taskLogs {
		a.IsTrue(log.TaskId == taskId)
		t.Log(log.Id, log.TaskId, log.Data)
	}

	if len(taskLogs) > 0 {
		t.Log("=======")
		taskLogs, err := AgentLogDAO().ListTaskLogs("test", taskId, taskLogs[len(taskLogs)-1].Id.Hex(), 2)
		if err != nil {
			t.Fatal(err)
		}

		for _, log := range taskLogs {
			a.IsTrue(log.TaskId == taskId)
			t.Log(log.Id, log.TaskId, log.Data)
		}
	}
}

func TestMongoAgentLogDAO_FindLatestTaskLog(t *testing.T) {
	a := assert.NewAssertion(t)

	taskId := "abc"
	taskLog, err := AgentLogDAO().FindLatestTaskLog("test", taskId)
	if err != nil {
		t.Fatal(err)
	}
	if taskLog == nil {
		t.Log("not found")
		return
	}
	a.IsTrue(taskLog.TaskId == taskId)
	t.Log(stringutil.JSONEncodePretty(taskLog))
}
