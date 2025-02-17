package hwctc

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"iptv/internal/app/iptv"
	"net/http"
	"regexp"
	"time"
)

// getLiveplayChannelProgramList 获取指定频道的节目单列表（sc）
func (c *Client) getLiveplayChannelProgramList(ctx context.Context, token *Token, channel *iptv.Channel) (*iptv.ChannelProgramList, error) {
	// 创建请求
	req, err := http.NewRequestWithContext(ctx, http.MethodGet,
		fmt.Sprintf("http://%s/EPG/jsp/liveplay_30/en/getTvodData.jsp", c.host), nil)
	if err != nil {
		return nil, err
	}

	// 增加请求参数
	params := req.URL.Query()
	params.Add("channelId", channel.ChannelID)
	req.URL.RawQuery = params.Encode()

	// 设置请求头
	c.setCommonHeaders(req)

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

	if resp.StatusCode == http.StatusNotFound {
		return nil, ErrEPGApiNotFound
	} else if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("http status code: %d", resp.StatusCode)
	}

	// 解析响应内容
	result, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// 正则提取JSON配置
	regex := regexp.MustCompile("parent.jsonBackLookStr = (.+?);")
	matches := regex.FindSubmatch(result)
	if len(matches) != 2 {
		return nil, ErrParseChProgList
	}

	// 解析节目单
	dateProgramList, err := parseLiveplayChannelProgramList(matches[1])
	if err != nil {
		return nil, err
	}

	return &iptv.ChannelProgramList{
		ChannelId:       channel.ChannelID,
		ChannelName:     channel.ChannelName,
		DateProgramList: dateProgramList,
	}, nil
}

// parseLiveplayChannelProgramList 解析频道节目单列表
func parseLiveplayChannelProgramList(rawData []byte) ([]iptv.DateProgram, error) {
	// 动态解析Json
	var rawArray []any
	err := json.Unmarshal(rawData, &rawArray)
	if err != nil {
		return nil, err
	}

	if len(rawArray) != 2 {
		return nil, ErrParseChProgList
	}

	dateProgList, ok := rawArray[1].([]any)
	if !ok {
		return nil, ErrParseChProgList
	} else if len(dateProgList) == 0 {
		return nil, ErrChProgListIsEmpty
	}

	// 遍历多个日期的节目单
	dateProgramList := make([]iptv.DateProgram, 0, len(dateProgList))
	for _, rawProgList := range dateProgList {
		progList, ok := rawProgList.([]any)
		if !ok {
			return nil, ErrParseChProgList
		} else if len(progList) == 0 {
			continue
		}

		// 遍历单个日期中的节目单
		programList := make([]iptv.Program, 0, len(progList))
		for _, rawProg := range progList {
			prog, ok := rawProg.(map[string]any)
			if !ok {
				return nil, ErrParseChProgList
			} else if len(prog) == 0 {
				continue
			}

			programName := prog["programName"].(string)
			beginTimeFormatStr := prog["beginTimeFormat"].(string)
			endTimeFormatStr := prog["endTimeFormat"].(string)
			startTimeStr := prog["startTime"].(string)
			endTimeStr := prog["endTime"].(string)

			if endTimeStr == "00:00" {
				// 临界值特殊处理
				endTimeStr = "23:59"

				// IPTV返回的结束时间为0点的节目单存在BUG，endTimeFormat错误设置为了当天的零点而不是第二天的零点
				// BUG数据示例：{"beginTimeFormat":"20241130232400","isPlayable":"0","programName":"典籍里的中国Ⅱ(6)","contentId":"755597800","index":"335","startTime":"23:24","endTime":"00:00","channelId":"658582938","endTimeFormat":"20241130000000"}
				if (beginTimeFormatStr[:8] + "000000") == endTimeFormatStr {
					endTimeFormat, err := time.ParseInLocation("20060102150405", endTimeFormatStr, time.Local)
					if err != nil {
						return nil, err
					}
					endTimeFormat = endTimeFormat.Add(24 * time.Hour)
					endTimeFormatStr = endTimeFormat.Format("20060102150405")
				}
			}

			programList = append(programList, iptv.Program{
				ProgramName:     programName,
				BeginTimeFormat: beginTimeFormatStr,
				EndTimeFormat:   endTimeFormatStr,
				StartTime:       startTimeStr,
				EndTime:         endTimeStr,
			})
		}

		beginTime, err := time.ParseInLocation("20060102150405", programList[0].BeginTimeFormat, time.Local)
		if err != nil {
			return nil, err
		}
		// 时间取整到天
		date := time.Date(beginTime.Year(), beginTime.Month(), beginTime.Day(), 0, 0, 0, 0, beginTime.Location())
		dateProgramList = append(dateProgramList, iptv.DateProgram{
			Date:        date,
			ProgramList: programList,
		})
	}
	return dateProgramList, nil
}
