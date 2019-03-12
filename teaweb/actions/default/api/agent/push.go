package agent

import (
	"context"
	"encoding/json"
	"github.com/TeaWeb/code/teaconfigs/agents"
	"github.com/TeaWeb/code/teaconfigs/notices"
	"github.com/TeaWeb/code/teamongo"
	"github.com/TeaWeb/code/teaweb/actions/default/agents/agentutils"
	"github.com/TeaWeb/code/teaweb/actions/default/notices/noticeutils"
	"github.com/iwind/TeaGo/actions"
	"github.com/iwind/TeaGo/logs"
	"github.com/iwind/TeaGo/maps"
	"github.com/iwind/TeaGo/types"
	"github.com/iwind/TeaGo/utils/time"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"io/ioutil"
	"strings"
	"sync"
	"time"
)

type PushAction actions.Action

// 接收推送的数据
func (this *PushAction) Run(params struct{}) {
	agent := this.Context.Get("agent").(*agents.AgentConfig)

	data, err := ioutil.ReadAll(this.Request.Body)
	if err != nil {
		logs.Error(err)
		this.Fail("read body error")
	}

	m := maps.Map{}
	err = json.Unmarshal(data, &m)
	if err != nil {
		logs.Error(err)
		this.Fail("unmarshal error")
	}

	timestamp := m.GetInt64("timestamp")
	t := time.Unix(timestamp, 0)

	eventDomain := m.GetString("event")

	if eventDomain == "ProcessEvent" { // 进程事件
		event := agentutils.ProcessLog{
			Id:         primitive.NewObjectID(),
			AgentId:    agent.Id,
			TaskId:     m.GetString("taskId"),
			ProcessId:  m.GetString("uniqueId"),
			ProcessPid: m.GetInt("pid"),
			EventType:  m.GetString("eventType"),
			Data:       m.GetString("data"),
			Timestamp:  timestamp,
			TimeFormat: struct {
				Year   string `bson:"year" json:"year"`
				Month  string `bson:"month" json:"month"`
				Day    string `bson:"day" json:"day"`
				Hour   string `bson:"hour" json:"hour"`
				Minute string `bson:"minute" json:"minute"`
				Second string `bson:"second" json:"second"`
			}{
				Year:   timeutil.Format("Y", t),
				Month:  timeutil.Format("Ym", t),
				Day:    timeutil.Format("Ymd", t),
				Hour:   timeutil.Format("YmdH", t),
				Minute: timeutil.Format("YmdHi", t),
				Second: timeutil.Format("YmdHis", t),
			},
		}

		coll := this.selectProcessEventCollection(agent.Id)
		_, err = coll.InsertOne(context.Background(), event)
		if err != nil {
			logs.Error(err)
		}
	} else if eventDomain == "ItemEvent" { // 监控项事件
		appId := m.GetString("appId")
		itemId := m.GetString("itemId")
		app := agentutils.FindAgentApp(agent, appId)
		if app == nil {
			this.Success()
		}

		item := app.FindItem(itemId)
		if item == nil {
			this.Success()
		}

		v := m.Get("value")
		threshold, level, message := item.TestValue(v)

		// 通知消息
		if level != notices.NoticeLevelNone {
			notice := notices.NewNotice()
			notice.SetTime(t)
			notice.Message = message
			notice.Agent = notices.AgentCond{
				AgentId: agent.Id,
				AppId:   appId,
				ItemId:  itemId,
				Level:   level,
			}
			if threshold != nil {
				notice.Agent.Threshold = threshold.Expression()
			}
			err = noticeutils.NewNoticeQuery().Insert(notice)
			if err != nil {
				logs.Error(err)
			}

			setting := notices.SharedNoticeSetting()

			// 通过媒介发送通知

			fullMessage := "消息：" + message + "\n时间：" + timeutil.Format("Y-m-d H:i:s", t)
			linkNames := []string{}
			for _, l := range agentutils.FindNoticeLinks(notice) {
				linkNames = append(linkNames, types.String(l["name"]))
			}
			if len(linkNames) > 0 {
				fullMessage += "\n位置：" + strings.Join(linkNames, "/")
			}

			// 查找分组，如果分组中有通知设置，则使用分组中的通知设置
			isNotified := false
			receiverIds := []string{}
			groupId := ""
			if len(agent.GroupIds) > 0 {
				groupId = agent.GroupIds[0]
			}
			group := agents.SharedGroupConfig().FindGroup(groupId)
			if group != nil {
				receivers, found := group.NoticeSetting[level]
				if found && len(receivers) > 0 {
					isNotified = true
					receiverIds = setting.NotifyReceivers(level, receivers, fullMessage, func(receiverId string, minutes int) int {
						return noticeutils.CountReceivedNotices(receiverId, map[string]interface{}{
							"agent.agentId": agent.Id,
							"agent.appId":   appId,
							"agent.itemId":  itemId,
						}, minutes)
					})
				}
			}

			// 全局通知
			if !isNotified {
				receiverIds = setting.Notify(level, fullMessage, func(receiverId string, minutes int) int {
					return noticeutils.CountReceivedNotices(receiverId, map[string]interface{}{
						"agent.agentId": agent.Id,
						"agent.appId":   appId,
						"agent.itemId":  itemId,
					}, minutes)
				})
			}
			if len(receiverIds) > 0 {
				noticeutils.UpdateNoticeReceivers(notice.Id, receiverIds)
			}
		}

		value := &agents.Value{
			Id:          primitive.NewObjectID(),
			AppId:       appId,
			AgentId:     agent.Id,
			ItemId:      itemId,
			Value:       v,
			Error:       m.GetString("error"),
			NoticeLevel: level,
			CreatedAt:   time.Now().Unix(),
		}
		if threshold != nil {
			value.Threshold = threshold.Expression()
		}
		value.SetTime(t)

		err := teamongo.NewAgentValueQuery().Insert(value)
		if err != nil {
			logs.Error(err)
		}
	} else if eventDomain == "SystemAppsEvent" {
		result := struct {
			Apps []*agents.AppConfig
		}{}
		err = json.Unmarshal(data, &result)
		if err != nil {
			logs.Error(err)
		} else {
			agentRuntime := agentutils.FindAgentRuntime(agent)
			agentRuntime.ResetSystemApps()
			agentRuntime.AddApps(result.Apps)
		}
	}

	this.Success()
}

var agentCollectionMap = map[string]*teamongo.Collection{} // agentId => collection
var agentCollectionLocker = sync.Mutex{}

func (this *PushAction) selectProcessEventCollection(agentId string) *teamongo.Collection {
	createdNew := false

	agentCollectionLocker.Lock()
	coll, found := agentCollectionMap[agentId]
	if !found {
		createdNew = true

		coll = teamongo.FindCollection("logs.agent." + agentId)
		agentCollectionMap[agentId] = coll
	}
	agentCollectionLocker.Unlock()

	if createdNew {
		coll.CreateIndex(map[string]bool{
			"agentId": true,
		})
		coll.CreateIndex(map[string]bool{
			"taskId": true,
		})
	}

	return coll
}
