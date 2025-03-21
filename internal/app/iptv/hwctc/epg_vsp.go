package hwctc

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"iptv/internal/app/iptv"
	"net/http"
	"slices"
	"strconv"
	"time"
)

type vspQueryChannel struct {
	ChannelIDs []int64 `json:"channelIDs"`
}

type vspQueryPlaybill struct {
	Type          string `json:"type"`
	StartTime     string `json:"startTime"`
	EndTime       string `json:"endTime"`
	Count         string `json:"count"`
	Offset        string `json:"offset"`
	IsFillProgram string `json:"isFillProgram"`
	MustIncluded  string `json:"mustIncluded"`
}

// vspQueryPayload 请求体
type vspQueryPayload struct {
	QueryChannel  *vspQueryChannel  `json:"queryChannel"`
	QueryPlaybill *vspQueryPlaybill `json:"queryPlaybill"`
	NeedChannel   string            `json:"needChannel"`
}

type vspResponseResult struct {
	RetMsg  string `json:"retMsg"`
	RetCode string `json:"retCode"`
}

type vspResponsePlaybillLiteRating struct {
	Name string `json:"name"`
	ID   string `json:"ID"`
}

type vspResponsePlaybillLite struct {
	Rating         *vspResponsePlaybillLiteRating `json:"rating"`
	IsNPVR         string                         `json:"isNPVR"`
	StartTime      string                         `json:"startTime"`
	ID             string                         `json:"ID"`
	ChannelID      string                         `json:"channelID"`
	CUTVStatus     string                         `json:"CUTVStatus"`
	IsFillProgram  string                         `json:"isFillProgram"`
	IsCPVR         string                         `json:"isCPVR"`
	Name           string                         `json:"name"`
	ReminderStatus string                         `json:"reminderStatus"`
	EndTime        string                         `json:"endTime"`
	IsCUTV         string                         `json:"isCUTV"`
}

type vspResponseChannelPlaybills struct {
	PlaybillCount string                    `json:"playbillCount"`
	PlaybillLites []vspResponsePlaybillLite `json:"playbillLites"`
}

// vspResponse 响应体
type vspResponse struct {
	Result           *vspResponseResult            `json:"result"`
	Total            string                        `json:"total"`
	ChannelPlaybills []vspResponseChannelPlaybills `json:"channelPlaybills"`
}

// getVspChannelProgramList 获取指定频道的节目单列表（hb）
func (c *Client) getVspChannelProgramList(ctx context.Context, token *Token, channel *iptv.Channel) (*iptv.ChannelProgramList, error) {
	// 获取未来一天的日期
	tomorrow := time.Now().AddDate(0, 0, 1)
	tomorrow = time.Date(tomorrow.Year(), tomorrow.Month(), tomorrow.Day(), 0, 0, 0, 0, tomorrow.Location())

	// 根据当前频道的时移范围，预估EPG的查询时间范围（加上未来一天）
	epgBackDay := int(channel.TimeShiftLength.Hours()/24) + 1
	// 限制EPG查询的最大时间范围
	if epgBackDay > maxBackDay {
		epgBackDay = maxBackDay
	}

	// 从未来一天开始往前，倒查多个日期的节目单
	dateProgramList := make([]iptv.DateProgram, 0, epgBackDay+1)
	for i := 0; i <= epgBackDay; i++ {
		// 获取起止时间
		startDate := tomorrow.AddDate(0, 0, -i)
		endDate := startDate.AddDate(0, 0, 1)

		// 获取指定日期的节目单列表
		programList, err := c.getVspChannelDateProgram(ctx, token, channel.ChannelID, startDate.UnixMilli(), endDate.UnixMilli(), 0)
		if err != nil {
			if errors.Is(err, ErrEPGApiNotFound) {
				return nil, err
			}
			c.logger.Sugar().Warnf("Failed to get the program list for channel %s on %s. Error: %v", channel.ChannelName, startDate.Format("20060102"), err)
			continue
		}

		dateProgramList = append(dateProgramList, iptv.DateProgram{
			Date:        startDate,
			ProgramList: programList,
		})
	}

	return &iptv.ChannelProgramList{
		ChannelId:       channel.ChannelID,
		ChannelName:     channel.ChannelName,
		DateProgramList: dateProgramList,
	}, nil
}

// getVspChannelDateProgram 获取指定频道的某日期的节目单列表
func (c *Client) getVspChannelDateProgram(ctx context.Context, token *Token, channelId string, startTime, endTime int64, offset int) ([]iptv.Program, error) {
	channelIdInt, err := strconv.ParseInt(channelId, 10, 64)
	if err != nil {
		return nil, err
	}

	// 创建请求体
	payload := &vspQueryPayload{
		QueryChannel: &vspQueryChannel{
			ChannelIDs: []int64{channelIdInt},
		},
		QueryPlaybill: &vspQueryPlaybill{
			Type:          "0",
			StartTime:     strconv.FormatInt(startTime, 10),
			EndTime:       strconv.FormatInt(endTime, 10),
			Count:         "100",
			Offset:        strconv.Itoa(offset),
			IsFillProgram: "0",
			MustIncluded:  "0",
		},
		NeedChannel: "0",
	}
	// 创建请求体bytes
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	// 创建请求
	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		fmt.Sprintf("http://%s/VSP/V3/QueryPlaybillList", c.host), bytes.NewReader(payloadBytes))
	if err != nil {
		return nil, err
	}

	// 设置请求头
	c.setCommonHeaders(req)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded; charset=UTF-8")
	req.Header.Set("X-Requested-With", "XMLHttpRequest")

	// 设置Cookie
	req.AddCookie(&http.Cookie{
		Name:  "JSESSIONID",
		Value: token.JSESSIONID,
	})

	// 执行请求
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound || resp.StatusCode >= http.StatusInternalServerError {
		return nil, ErrEPGApiNotFound
	} else if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("http status code: %d", resp.StatusCode)
	}

	// 解析响应内容
	var response vspResponse
	if err = json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("parse response failed: %w", err)
	} else if response.Result == nil || response.Result.RetCode != "000000000" || len(response.ChannelPlaybills) == 0 {
		// 调用失败
		return nil, fmt.Errorf("the API returned failed, response: %+v", response)
	}

	// 解析节目单
	channelPlaybills := response.ChannelPlaybills[0]
	programList, err := parseVspChannelDateProgram(channelPlaybills.PlaybillLites)
	if err != nil {
		return nil, err
	}

	// 若有分页则进行递归调用
	count, err := strconv.Atoi(channelPlaybills.PlaybillCount)
	if err != nil {
		return nil, err
	}
	if count > (offset + 100) {
		nextProgramList, err := c.getVspChannelDateProgram(ctx, token, channelId, startTime, endTime, offset+100)
		if err != nil {
			return nil, err
		}
		programList = slices.Concat(programList, nextProgramList)
	}

	return programList, nil
}

// parseVspChannelDateProgram 解析频道节目单列表
func parseVspChannelDateProgram(playbillLites []vspResponsePlaybillLite) ([]iptv.Program, error) {
	if len(playbillLites) == 0 {
		return nil, ErrChProgListIsEmpty
	}

	// 遍历单个日期中的节目单
	programList := make([]iptv.Program, 0, len(playbillLites))
	for _, playbillLite := range playbillLites {
		startTimeInt, err := strconv.ParseInt(playbillLite.StartTime, 10, 64)
		if err != nil {
			return nil, err
		}
		endTimeInt, err := strconv.ParseInt(playbillLite.EndTime, 10, 64)
		if err != nil {
			return nil, err
		}

		// 时间戳转换
		bTime := time.UnixMilli(startTimeInt)
		eTime := time.UnixMilli(endTimeInt)

		// 临界值特殊处理
		endTimeStr := eTime.Format("15:04")
		if endTimeStr == "00:00" {
			endTimeStr = "23:59"
		}

		programList = append(programList, iptv.Program{
			ProgramName:     playbillLite.Name,
			BeginTimeFormat: bTime.Format("20060102150405"),
			EndTimeFormat:   eTime.Format("20060102150405"),
			StartTime:       bTime.Format("15:04"),
			EndTime:         endTimeStr,
		})
	}
	return programList, nil
}
