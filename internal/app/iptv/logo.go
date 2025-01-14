package iptv

import (
	"regexp"
	"strconv"
	"strings"
)

const logoDirName = "logos"

type ChannelLogoRule struct {
	Name string
	Rule *regexp.Regexp
}

func (l *ChannelLogoRule) ResolveName(matches []string) string {
	s := l.Name
	if len(matches) > 1 {
		for i, ma := range matches[1:] {
			s = strings.ReplaceAll(s, "$G"+strconv.FormatInt(int64(i+1), 10), ma)
		}
	}
	return s
}

// GetChannelLogoName 根据频道名称识别频道台标logo
func GetChannelLogoName(chLogoRuleList []ChannelLogoRule, channelName string) string {
	for _, chLogoRule := range chLogoRuleList {
		matches := chLogoRule.Rule.FindStringSubmatch(channelName)
		if len(matches) > 0 {
			return chLogoRule.ResolveName(matches)
		}
	}
	return channelName
}
