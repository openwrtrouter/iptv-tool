package iptv

import (
	"time"
)

// ChannelProgramList 频道节目单列表
type ChannelProgramList struct {
	ChannelId       string        `json:"channelId"`             // 频道Id
	ChannelName     string        `json:"channelName,omitempty"` // 频道名称
	DateProgramList []DateProgram `json:"dateProgramList"`       // 不同日期的频道列表
}

// DateProgram 一天的节目单列表
type DateProgram struct {
	Date        time.Time `json:"date"`
	ProgramList []Program `json:"programList"`
}

// Program 节目单
type Program struct {
	ProgramName     string `json:"programName"`     // 节目名称
	BeginTimeFormat string `json:"beginTimeFormat"` // 格式化的开始时间，例如：20241122205700
	EndTimeFormat   string `json:"endTimeFormat"`   // 格式化的结束时间，例如：20241122210100
	StartTime       string `json:"startTime"`       // 开始时间，例如：20:57
	EndTime         string `json:"endTime"`         // 结束时间，例如：21:01
}
