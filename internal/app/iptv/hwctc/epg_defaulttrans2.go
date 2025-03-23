package hwctc

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"iptv/internal/app/iptv"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

type defaulttrans2Respone struct {
	Data  []defaulttrans2ChannelProg `json:"data"`
	Title []string                   `json:"title"`
}

type defaulttrans2ChannelProg struct {
	ProgName    string `json:"progName"`
	ScrollFlag  int    `json:"scrollFlag"`
	StartTime   string `json:"startTime"`
	EndTime     string `json:"endTime"`
	SubProgName string `json:"subProgName"`
	State       string `json:"state"`
	ProgId      string `json:"progId"`
}

// getDefaulttrans2ChannelProgramList 获取指定频道的节目单列表（sd）
func (c *Client) getDefaulttrans2ChannelProgramList(ctx context.Context, token *Token, channel *iptv.Channel) (*iptv.ChannelProgramList, error) {
	now := time.Now()
	now = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	// 从当天开始往前，倒查多个日期的节目单
	dateSize := 7
	dateProgramList := make([]iptv.DateProgram, 0, dateSize)
	for i := 0; i < dateSize; i++ {
		date := now.AddDate(0, 0, -i)

		// 获取指定日期的节目单列表
		programList, chDateSize, err := c.getDefaulttrans2ChannelDateProgram(ctx, token, channel, date, -i)
		if err != nil {
			if errors.Is(err, ErrEPGApiNotFound) {
				return nil, err
			}
			c.logger.Sugar().Warnf("Failed to get the program list for channel %s on %s (index: %d). Error: %v", channel.ChannelName, date.Format("20060102"), -i, err)
			continue
		}

		if i == 0 {
			dateSize = chDateSize
		}
		dateProgramList = append(dateProgramList, iptv.DateProgram{
			Date:        date,
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
func (c *Client) getDefaulttrans2ChannelDateProgram(ctx context.Context, token *Token, channel *iptv.Channel, date time.Time, index int) ([]iptv.Program, int, error) {
	// 创建请求
	req, err := http.NewRequestWithContext(ctx, http.MethodGet,
		fmt.Sprintf("http://%s/EPG/jsp/defaulttrans2/en/datajsp/getTvodProgListByIndex.jsp", c.host), nil)
	if err != nil {
		return nil, 0, err
	}

	// 增加请求参数
	params := req.URL.Query()
	params.Add("CHANNELID", channel.ChannelID)
	params.Add("index", strconv.Itoa(index))
	req.URL.RawQuery = params.Encode()

	// 设置请求头
	c.setCommonHeaders(req)
	req.Header.Set("Referer", fmt.Sprintf("http://%s/EPG/jsp/defaulttrans2/en/chanMiniList.html", c.host))

	// 设置Cookie
	cookies := []*http.Cookie{
		{Name: "maidianFlag", Value: "1"},
		{Name: "navNameFocus", Value: "3"},
		{Name: "jumpTime", Value: "0"},
		{Name: "channelTip", Value: "1"},
		{Name: "lastChanNum", Value: "1"},
		{Name: "STARV_TIMESHFTCID", Value: channel.ChannelID},
		{Name: "STARV_TIMESHFTCNAME", Value: url.QueryEscape(channel.ChannelName)},
		{Name: "JSESSIONID", Value: token.JSESSIONID},
	}
	for _, cookie := range cookies {
		req.AddCookie(cookie)
	}

	// 执行请求
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound || resp.StatusCode >= http.StatusInternalServerError {
		return nil, 0, ErrEPGApiNotFound
	} else if resp.StatusCode != http.StatusOK {
		return nil, 0, fmt.Errorf("http status code: %d", resp.StatusCode)
	}

	// 解析响应内容
	var response defaulttrans2Respone
	if err = json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, 0, fmt.Errorf("parse response failed: %w", err)
	}

	// 解析节目单信息
	return parseDefaulttrans2ChannelDateProgram(response, date, index)
}

// parseDefaulttrans2ChannelDateProgram 解析频道节目单列表
func parseDefaulttrans2ChannelDateProgram(response defaulttrans2Respone, date time.Time, index int) ([]iptv.Program, int, error) {
	if len(response.Data) == 0 {
		return nil, 0, ErrChProgListIsEmpty
	} else if len(response.Title) == 0 {
		return nil, 0, fmt.Errorf("no date title list")
	}

	// 比较日期是否正确
	datePos := len(response.Title) - 1 + index
	if datePos >= len(response.Title) || datePos < 0 {
		return nil, 0, fmt.Errorf("invalid date position: %d", datePos)
	} else if !strings.HasPrefix(response.Title[datePos], date.Format("02")) {
		return nil, 0, fmt.Errorf("the program date does not match the query date")
	}

	dateStr := date.Format("20060102")
	// 遍历单个日期中的节目单
	programList := make([]iptv.Program, 0, len(response.Data))
	for i, prog := range response.Data {
		// 处理节目单的开始和结束时间
		startTimeStr := prog.StartTime
		if i == 0 {
			// 将第一个节目单的开始时间统一设置为零点
			startTimeStr = "00:00"
		} else if len(startTimeStr) > 5 {
			startTimeStr = startTimeStr[:5]
		}
		endTimeStr := prog.EndTime
		if len(endTimeStr) > 5 {
			endTimeStr = endTimeStr[:5]
		}

		bTime, err := time.ParseInLocation("20060102 15:04", dateStr+" "+startTimeStr, time.Local)
		if err != nil {
			return nil, 0, err
		}
		eTime, err := time.ParseInLocation("20060102 15:04", dateStr+" "+endTimeStr, time.Local)
		if err != nil {
			return nil, 0, err
		}
		// 处理跨天的节目单数据，将结束时间改为第二天的零点
		if bTime.After(eTime) {
			tempDate := date.AddDate(0, 0, 1)
			eTime = time.Date(tempDate.Year(), tempDate.Month(), tempDate.Day(), 0, 0, 0, 0, tempDate.Location())
			endTimeStr = "23:59"
		}

		// 组装节目单对象
		programList = append(programList, iptv.Program{
			ProgramName:     prog.ProgName,
			BeginTimeFormat: bTime.Format("20060102150405"),
			EndTimeFormat:   eTime.Format("20060102150405"),
			StartTime:       startTimeStr,
			EndTime:         endTimeStr,
		})
		// 丢弃后续第二天的节目单数据，如果存在的话
		if endTimeStr == "23:59" {
			break
		}
	}
	return programList, len(response.Title), nil
}
