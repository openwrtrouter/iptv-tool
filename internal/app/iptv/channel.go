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
)

type Channel struct {
	ChannelID       string        `json:"channelID"`       // 频道ID
	ChannelName     string        `json:"channelName"`     // 频道名称
	ChannelURL      *url.URL      `json:"channelURL"`      // 频道URL
	TimeShift       string        `json:"timeShift"`       // 时移类型
	TimeShiftLength time.Duration `json:"timeShiftLength"` // 支持的时移长度
	TimeShiftURL    string        `json:"timeShiftURL"`    // 时移地址（回放地址）
}

// GetChannelList 获取频道列表
func (c *Client) GetChannelList(ctx context.Context, token *Token) ([]Channel, error) {
	// 计算JSESSIONID的MD5
	hash := md5.Sum([]byte(token.JSESSIONID))
	// 转换为16进制字符串并转换为大写，即为tempKey
	tempKey := hex.EncodeToString(hash[:])

	// 组装请求数据
	data := map[string]string{
		"conntype":  "dhcp",
		"UserToken": token.UserToken,
		"tempKey":   tempKey,
		"stbid":     token.Stbid,
		"SupportHD": "1",
		"UserID":    c.config.UserID,
		"Lang":      "1",
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
	chRegex := regexp.MustCompile("ChannelID=\"(.+?)\",ChannelName=\"(.+?)\",UserChannelID=\".+?\",ChannelURL=\"(.+?)\",TimeShift=\"(.+?)\",TimeShiftLength=\"(\\d+?)\".+?,TimeShiftURL=\"(.+?)\"")
	matchesList := chRegex.FindAllSubmatch(result, -1)
	if matchesList == nil {
		return nil, fmt.Errorf("failed to extract channel list")
	}

	// 过滤掉特殊频道的正则表达式
	chExcludeRegex := regexp.MustCompile("^(.+?(画中画|单音轨))|(\\d+)$")

	channels := make([]Channel, 0, len(matchesList))
	for _, matches := range matchesList {
		if len(matches) != 7 {
			continue
		}

		channelName := string(matches[2])
		// 过滤掉特殊频道
		if chExcludeRegex.MatchString(channelName) {
			continue
		}

		// channelURL类型转换
		channelURL, err := url.Parse(string(matches[3]))
		if err != nil {
			continue
		}

		// TimeShiftLength类型转换
		timeShiftLength, err := strconv.ParseInt(string(matches[5]), 10, 64)
		if err != nil {
			continue
		}

		channels = append(channels, Channel{
			ChannelID:       string(matches[1]),
			ChannelName:     channelName,
			ChannelURL:      channelURL,
			TimeShift:       string(matches[4]),
			TimeShiftLength: time.Duration(timeShiftLength) * time.Minute,
			TimeShiftURL:    string(matches[6]),
		})
	}
	return channels, nil
}

// ToM3UFormat 转换为M3U格式内容
func ToM3UFormat(channels []Channel, udpxyURL string) (string, error) {
	if len(channels) == 0 {
		return "", errors.New("no channels found")
	}

	var sb strings.Builder
	sb.WriteString("#EXTM3U\n")
	for _, channel := range channels {
		var err error
		var channelURL string
		if udpxyURL != "" {
			channelURL, err = url.JoinPath(udpxyURL, fmt.Sprintf("/rtp/%s", channel.ChannelURL.Host))
			if err != nil {
				return "", err
			}
		} else {
			channelURL = channel.ChannelURL.String()
		}
		m3uLine := fmt.Sprintf("#EXTINF:-1 ,%s\n%s\n",
			channel.ChannelName, channelURL)
		sb.WriteString(m3uLine)
	}
	return sb.String(), nil
}

// ToTxtFormat 转换为txt格式内容
func ToTxtFormat(channels []Channel, udpxyURL string) (string, error) {
	if len(channels) == 0 {
		return "", errors.New("no channels found")
	}

	var sb strings.Builder
	for _, channel := range channels {
		var err error
		var channelURL string
		if udpxyURL != "" {
			channelURL, err = url.JoinPath(udpxyURL, fmt.Sprintf("/rtp/%s", channel.ChannelURL.Host))
			if err != nil {
				return "", err
			}
		} else {
			channelURL = channel.ChannelURL.String()
		}

		m3uLine := fmt.Sprintf("%s,%s\n",
			channel.ChannelName, channelURL)
		sb.WriteString(m3uLine)
	}
	return sb.String(), nil
}
