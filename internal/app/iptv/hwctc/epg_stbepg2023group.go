package hwctc

import (
	"context"
	"encoding/json"
	"fmt"
	"iptv/internal/app/iptv"
	"iptv/internal/pkg/util"
	"net/http"
	"net/url"
	"slices"
	"strconv"
	"strings"
	"time"
)

type stbEpg2023GroupResponse[T any] struct {
	Data    T      `json:"data"`
	ErrCode string `json:"errCode"`
	ErrMsg  string `json:"errMsg"`
	Status  string `json:"status"`
}

type stbEpg2023GroupCategory struct {
	Name string `json:"name"`
	ID   string `json:"id"`
}

type stbEpg2023GroupChannel struct {
	AuthCode string `json:"authCode"`
	Code     string `json:"code"`
	Name     string `json:"name"`
	IsCharge string `json:"isCharge"`
	ID       string `json:"ID"`
	MixNo    string `json:"mixNo"`
	MediaID  string `json:"mediaID"`
}

type stbEpg2023GroupChannelProg struct {
	Name      string `json:"name"`
	StartTime int64  `json:"startTime"`
	ID        string `json:"ID"`
	EndTime   int64  `json:"endTime"`
	ChannelID string `json:"channelID"`
	Status    string `json:"status"`
}

// getStbEpg2023GroupAllChannelProgramList 获取全部频道的节目单列表（fj）
func (c *Client) getStbEpg2023GroupAllChannelProgramList(ctx context.Context, channels []iptv.Channel, token *Token) ([]iptv.ChannelProgramList, error) {
	// 获取“全部”类别的ID
	categoryID, err := c.getStbEpg2023GroupChannelCategoryID(ctx, "全部", token)
	if err != nil {
		return nil, err
	}

	// 获取所有频道列表的code
	stbEpg2023GrouChList, err := c.getStbEpg2023GroupChannelList(ctx, categoryID, token)
	if err != nil {
		return nil, err
	}
	// 获取频道ID和频道Code的映射关系
	chIdCodeMap := make(map[string]string, len(stbEpg2023GrouChList))
	for _, stbEpg2023GrouCh := range stbEpg2023GrouChList {
		chIdCodeMap[stbEpg2023GrouCh.ID] = stbEpg2023GrouCh.Code
	}

	epg := make([]iptv.ChannelProgramList, 0, len(channels))
	for _, channel := range channels {
		// 跳过不支持回看的频道
		if channel.TimeShift != "1" || channel.TimeShiftLength <= 0 {
			continue
		}

		chCode, ok := chIdCodeMap[channel.ChannelID]
		if !ok {
			c.logger.Sugar().Warnf("Failed to get the code for channel %s.", channel.ChannelName)
			continue
		}

		// 获取单个频道的全部节目单列表
		progList, err := c.getStbEpg2023GroupChannelProgramList(ctx, token, &channel, chCode)
		if err != nil {
			c.logger.Sugar().Warnf("Failed to get the program list for channel %s. Error: %v", channel.ChannelName, err)
			continue
		}

		if progList != nil && len(progList.DateProgramList) > 0 {
			// 对频道的节目单按日期升序排序
			slices.SortFunc(progList.DateProgramList, func(a, b iptv.DateProgram) int {
				return a.Date.Compare(b.Date)
			})

			epg = append(epg, *progList)
		}
	}
	return epg, nil
}

// getChannelCate 获取指定频道类别的ID
func (c *Client) getStbEpg2023GroupChannelCategoryID(ctx context.Context, categoryName string, token *Token) (string, error) {
	// 组装请求数据
	data := map[string]string{
		"action": "getChannelCate",
	}
	body := url.Values{}
	for k, v := range data {
		body.Add(k, v)
	}

	// 创建请求
	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		fmt.Sprintf("http://%s/EPG/jsp/StbEpg2023Group/en/function/ajax/epg7getProperties.jsp", c.host), strings.NewReader(body.Encode()))
	if err != nil {
		return "", err
	}

	// 设置请求头
	c.setCommonHeaders(req)
	req.Header.Set("VIS-AJAX", "AjaxHttpRequest")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// 设置Cookie
	req.AddCookie(&http.Cookie{
		Name:  "JSESSIONID",
		Value: token.JSESSIONID,
	})

	// 执行请求
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound || resp.StatusCode >= http.StatusInternalServerError {
		return "", ErrEPGApiNotFound
	} else if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("http status code: %d", resp.StatusCode)
	}

	// 解析响应内容
	var response stbEpg2023GroupResponse[[]stbEpg2023GroupCategory]
	if err = json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return "", fmt.Errorf("parse response failed: %w", err)
	} else if response.Status != "1" {
		// 调用失败
		return "", fmt.Errorf("the API returned failed, errMsg: %s", response.ErrMsg)
	} else if len(response.Data) == 0 {
		// 未获取到频道分类信息
		return "", fmt.Errorf("no channel categories")
	}

	var categoryID string
	for _, category := range response.Data {
		if categoryName == category.Name {
			categoryID = category.ID
			break
		}
	}
	if categoryID == "" {
		return "", fmt.Errorf("channel category not found")
	}
	return categoryID, nil
}

// getStbEpg2023GroupChannelList 获取指定频道类别的频道列表
func (c *Client) getStbEpg2023GroupChannelList(ctx context.Context, categoryID string, token *Token) ([]stbEpg2023GroupChannel, error) {
	// 组装请求数据
	data := map[string]string{
		"action": "getChannelList",
		"cateID": categoryID,
		"type":   "",
	}
	body := url.Values{}
	for k, v := range data {
		body.Add(k, v)
	}

	// 创建请求
	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		fmt.Sprintf("http://%s/EPG/jsp/StbEpg2023Group/en/function/ajax/epg7getChannelByAjax.jsp", c.host), strings.NewReader(body.Encode()))
	if err != nil {
		return nil, err
	}

	// 设置请求头
	c.setCommonHeaders(req)
	req.Header.Set("VIS-AJAX", "AjaxHttpRequest")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

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

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("http status code: %d", resp.StatusCode)
	}

	// 解析响应内容
	var response stbEpg2023GroupResponse[[]stbEpg2023GroupChannel]
	if err = json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("parse response failed: %w", err)
	} else if response.Status != "1" {
		// 调用失败
		return nil, fmt.Errorf("the API returned failed, errMsg: %s", response.ErrMsg)
	} else if len(response.Data) == 0 {
		// 未找到频道列表
		return nil, fmt.Errorf("no channel list")
	}

	return response.Data, nil
}

// getStbEpg2023GroupChannelProgramList 获取指定频道的节目单列表
func (c *Client) getStbEpg2023GroupChannelProgramList(ctx context.Context, token *Token, channel *iptv.Channel, chCode string) (*iptv.ChannelProgramList, error) {
	// 根据当前频道的时移范围，预估EPG的查询时间范围（加上未来一天）
	epgBackDay := int(channel.TimeShiftLength.Hours()/24) + 1
	// 限制EPG查询的最大时间范围
	if epgBackDay > maxBackDay {
		epgBackDay = maxBackDay
	}

	// 计算开始、结束时间
	tomorrow := time.Now().AddDate(0, 0, 1)
	old := tomorrow.AddDate(0, 0, -epgBackDay)
	startTime := time.Date(old.Year(), old.Month(), old.Day(), 0, 0, 1, 534, tomorrow.Location()).UnixMilli()
	endTime := time.Date(tomorrow.Year(), tomorrow.Month(), tomorrow.Day(), 23, 59, 59, 534, tomorrow.Location()).UnixMilli()

	// 组装请求数据
	data := map[string]string{
		"action":    "getChannelProg",
		"code":      chCode,
		"channelID": channel.ChannelID,
		"endTime":   strconv.FormatInt(endTime, 10),
		"startTime": strconv.FormatInt(startTime, 10),
		"offset":    "0",
		"limit":     "2000",
	}
	body := url.Values{}
	for k, v := range data {
		body.Add(k, v)
	}

	// 创建请求
	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		fmt.Sprintf("http://%s/EPG/jsp/StbEpg2023Group/en/function/ajax/epg7getChannelByAjax.jsp", c.host), strings.NewReader(body.Encode()))
	if err != nil {
		return nil, err
	}

	// 设置请求头
	c.setCommonHeaders(req)
	req.Header.Set("VIS-AJAX", "AjaxHttpRequest")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

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

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("http status code: %d", resp.StatusCode)
	}

	// 解析响应内容
	var response stbEpg2023GroupResponse[[]stbEpg2023GroupChannelProg]
	if err = json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("parse response failed: %w", err)
	} else if response.Status != "1" {
		// 调用失败
		return nil, fmt.Errorf("the API returned failed, errMsg: %s", response.ErrMsg)
	}

	// 解析节目单
	dateProgramList, err := parseStbEpg2023GroupDateProgramList(response.Data)
	if err != nil {
		return nil, err
	}

	return &iptv.ChannelProgramList{
		ChannelId:       channel.ChannelID,
		ChannelName:     channel.ChannelName,
		DateProgramList: dateProgramList,
	}, nil
}

// parseStbEpg2023GroupDateProgramList 解析频道节目单列表
func parseStbEpg2023GroupDateProgramList(channelProgList []stbEpg2023GroupChannelProg) ([]iptv.DateProgram, error) {
	if len(channelProgList) == 0 {
		return nil, ErrChProgListIsEmpty
	}

	// 遍历频道节目单列表
	progMap := make(map[string][]iptv.Program)
	for _, channelProg := range channelProgList {
		// 时间戳转换
		bTime := time.UnixMilli(channelProg.StartTime)
		eTime := time.UnixMilli(channelProg.EndTime)

		// 临界值特殊处理
		endTimeStr := eTime.Format("15:04")
		if endTimeStr == "00:00" {
			endTimeStr = "23:59"
		}

		dateStr := bTime.Format("20060102")
		programList, ok := progMap[dateStr]
		if !ok {
			programList = make([]iptv.Program, 0)
		}
		programList = append(programList, iptv.Program{
			ProgramName:     channelProg.Name,
			BeginTimeFormat: bTime.Format("20060102150405"),
			EndTimeFormat:   eTime.Format("20060102150405"),
			StartTime:       bTime.Format("15:04"),
			EndTime:         endTimeStr,
		})
		progMap[dateStr] = programList
	}

	// 组装结果
	dateProgramList := make([]iptv.DateProgram, 0)
	for _, dateStr := range util.SortedMapKeys(progMap) {
		programList := progMap[dateStr]

		date, err := time.ParseInLocation("20060102", dateStr, time.Local)
		if err != nil {
			return nil, err
		}
		dateProgramList = append(dateProgramList, iptv.DateProgram{
			Date:        date,
			ProgramList: programList,
		})
	}
	return dateProgramList, nil
}
