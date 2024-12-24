package iptv

import (
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"
)

const SCHEME_IGMP = "igmp"

// Channel 频道信息
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
		if udpxyURL != "" && channel.ChannelURL.Scheme == SCHEME_IGMP {
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
			if udpxyURL != "" && channel.ChannelURL.Scheme == SCHEME_IGMP {
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
