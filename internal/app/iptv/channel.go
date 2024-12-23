package iptv

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"go.uber.org/zap"
)

type Channel struct {
	ChannelID       string        `json:"channelID"`       // 频道ID
	ChannelName     string        `json:"channelName"`     // 频道名称
	UserChannelID   string        `json:"userChannelID"`   // 频道号
	ChannelURL      *url.URL      `json:"channelURL"`      // 频道URL
	TimeShift       string        `json:"timeShift"`       // 时移类型
	TimeShiftLength time.Duration `json:"timeShiftLength"` // 支持的时移长度
	TimeShiftURL    *url.URL      `json:"timeShiftURL"`    // 时移地址（回放地址）

	GroupName string `json:"groupName"` // 程序识别的频道分类
}

// GetChannelList 获取频道列表
func (c *Client) GetChannelList(ctx context.Context, token *Token) ([]Channel, error) {
	// 计算JSESSIONID的MD5
	hash := md5.Sum([]byte(token.JSESSIONID))
	// 转换为16进制字符串并转换为大写，即为tempKey
	tempKey := hex.EncodeToString(hash[:])

	// 组装请求数据
	data := map[string]string{
		"conntype":  c.config.Conntype,
		"UserToken": token.UserToken,
		"tempKey":   tempKey,
		"stbid":     token.Stbid,
		"SupportHD": "1",
		"UserID":    c.config.UserID,
		"Lang":      c.config.Lang,
	}
	body := url.Values{}
	for k, v := range data {
		body.Add(k, v)
	}

	// 创建请求
	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		fmt.Sprintf("http://%s/EPG/jsp/getchannellistHWCTC.jsp", c.host), strings.NewReader(body.Encode()))
	if err != nil {
		return nil, err
	}

	// 设置请求头
	c.setCommonHeaders(req)
	req.Header.Set("Referer", fmt.Sprintf("http://%s/EPG/jsp/ValidAuthenticationHWCTC.jsp", c.host))
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
	result, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	chRegex := regexp.MustCompile("ChannelID=\"(.+?)\",ChannelName=\"(.+?)\",UserChannelID=\"(.+?)\",ChannelURL=\"(.+?)\",TimeShift=\"(.+?)\",TimeShiftLength=\"(\\d+?)\".+?,TimeShiftURL=\"(.+?)\"")
	matchesList := chRegex.FindAllSubmatch(result, -1)
	if matchesList == nil {
		return nil, fmt.Errorf("failed to extract channel list")
	}

	// 过滤掉特殊频道的正则表达式
	chExcludeRegex := regexp.MustCompile("^.+?(画中画|单音轨)$")

	channels := make([]Channel, 0, len(matchesList))
	for _, matches := range matchesList {
		if len(matches) != 8 {
			continue
		}

		channelName := string(matches[2])
		// 过滤掉特殊频道
		if chExcludeRegex.MatchString(channelName) {
			c.logger.Warn("This is not a normal channel, skip it.", zap.String("channelName", channelName))
			continue
		}

		// channelURL类型转换
		// channelURL可能同时返回组播和单播多个地址（通过|分割），这里优先取组播地址
		var channelURL *url.URL
		channelURLStrList := strings.Split(string(matches[4]), "|")
		for _, channelURLStr := range channelURLStrList {
			channelURL, err = url.Parse(channelURLStr)
			if err != nil {
				continue
			}

			if channelURL != nil && channelURL.Scheme == "igmp" {
				break
			}
		}

		if channelURL == nil {
			c.logger.Warn("The channelURL of this channel is illegal, skip it.", zap.String("channelName", channelName), zap.String("channelURL", string(matches[4])))
			continue
		}

		// TimeShiftLength类型转换
		timeShiftLength, err := strconv.ParseInt(string(matches[6]), 10, 64)
		if err != nil {
			c.logger.Warn("The timeShiftLength of this channel is illegal, skip it.", zap.String("channelName", channelName), zap.String("timeShiftLength", string(matches[6])))
			continue
		}

		// 解析时移地址
		timeShiftURL, err := url.Parse(string(matches[7]))
		if err != nil {
			c.logger.Warn("The timeShiftURL of this channel is illegal, skip it.", zap.String("channelName", channelName), zap.String("timeShiftURL", string(matches[7])))
			continue
		}
		// 重置时移地址的查询参数
		timeShiftURL.RawQuery = ""

		// 自动识别频道的分类
		groupName := getChannelGroupName(channelName)

		channels = append(channels, Channel{
			ChannelID:       string(matches[1]),
			ChannelName:     channelName,
			UserChannelID:   string(matches[3]),
			ChannelURL:      channelURL,
			TimeShift:       string(matches[5]),
			TimeShiftLength: time.Duration(timeShiftLength) * time.Minute,
			TimeShiftURL:    timeShiftURL,
			GroupName:       groupName,
		})
	}
	return channels, nil
}

// ToM3UFormat 转换为M3U格式内容
func ToM3UFormat(channels []Channel, udpxyURL, catchupSource string) (string, error) {
	if len(channels) == 0 {
		return "", errors.New("no channels found")
	}

	var sb strings.Builder
	sb.WriteString("#EXTM3U\n")
	for _, channel := range channels {
		var err error
		var channelURL string
		if udpxyURL != "" && channel.ChannelURL.Scheme == "igmp" {
			channelURL, err = url.JoinPath(udpxyURL, fmt.Sprintf("/rtp/%s", channel.ChannelURL.Host))
			if err != nil {
				return "", err
			}
		} else {
			channelURL = channel.ChannelURL.String()
		}
		var m3uLine string
		if channel.TimeShift == "1" && channel.TimeShiftLength > 0 {
			m3uLine = fmt.Sprintf("#EXTINF:-1 tvg-id=\"%s\" tvg-chno=\"%s\" catchup=\"%s\" catchup-source=\"%s\" catchup-days=\"%d\" group-title=\"%s\",%s\n%s\n",
				channel.ChannelID, channel.UserChannelID, "default", channel.TimeShiftURL.String()+catchupSource,
				int64(channel.TimeShiftLength.Hours()/24), channel.GroupName, channel.ChannelName, channelURL)
		} else {
			m3uLine = fmt.Sprintf("#EXTINF:-1 tvg-id=\"%s\" tvg-chno=\"%s\" group-title=\"%s\",%s\n%s\n",
				channel.ChannelID, channel.UserChannelID, channel.GroupName, channel.ChannelName, channelURL)
		}
		sb.WriteString(m3uLine)
	}
	return sb.String(), nil
}

// ToTxtFormat 转换为txt格式内容
func ToTxtFormat(channels []Channel, udpxyURL string) (string, error) {
	if len(channels) == 0 {
		return "", errors.New("no channels found")
	}

	// 对频道列表，按分组名称进行分组
	groupNames := make([]string, 0)
	groupChannelMap := make(map[string][]Channel)
	for _, channel := range channels {
		groupChannels, ok := groupChannelMap[channel.GroupName]
		if !ok {
			groupNames = append(groupNames, channel.GroupName)
			groupChannelMap[channel.GroupName] = []Channel{channel}
			continue
		}

		groupChannels = append(groupChannels, channel)
		groupChannelMap[channel.GroupName] = groupChannels
	}

	var sb strings.Builder
	// 为保证顺序，单独遍历分组名称的slices
	for _, groupName := range groupNames {
		groupChannels := groupChannelMap[groupName]

		// 输出分组信息
		groupLine := fmt.Sprintf("%s,#genre#\n", groupName)
		sb.WriteString(groupLine)

		// 输出频道信息
		for _, channel := range groupChannels {
			var err error
			var channelURL string
			if udpxyURL != "" && channel.ChannelURL.Scheme == "igmp" {
				channelURL, err = url.JoinPath(udpxyURL, fmt.Sprintf("/rtp/%s", channel.ChannelURL.Host))
				if err != nil {
					return "", err
				}
			} else {
				channelURL = channel.ChannelURL.String()
			}

			txtLine := fmt.Sprintf("%s,%s\n",
				channel.ChannelName, channelURL)
			sb.WriteString(txtLine)
		}
	}
	return sb.String(), nil
}
