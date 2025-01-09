package iptv

import (
	"regexp"
)

const otherChGroupName = "其他"

type ChannelGroupRules struct {
	Name  string           // 分组名称
	Rules []*regexp.Regexp // 分组规则
}

// GetChannelGroupName 根据频道名称自动获取分组名称
func GetChannelGroupName(chGroupRulesList []ChannelGroupRules, channelName string) string {
	// 自动识别频道的分类
	for _, chGroupRules := range chGroupRulesList {
		for _, groupRule := range chGroupRules.Rules {
			if groupRule.MatchString(channelName) {
				return chGroupRules.Name
			}
		}
	}
	return otherChGroupName
}
