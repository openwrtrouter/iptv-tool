package iptv

import (
	"regexp"
)

const otherGroupName = "其他"

// chGroupRuleMap 频道分组规则
var chGroupRuleMap = map[string][]*regexp.Regexp{
	"央视": {
		regexp.MustCompile("^(CCTV|中央).+?$"),
	},
	"卫视": {
		regexp.MustCompile("^[^(热门)].+?卫视.*?$"),
	},
	"国际": {
		regexp.MustCompile("^(CGTN|凤凰).+?$"),
	},
	"地方": {
		regexp.MustCompile("^(SCTV|CDTV).+?$"), // 四川
		regexp.MustCompile("^(浙江|杭州|民生|钱江|教科影视|好易购|西湖|青少体育).+?$"), // 浙江
	},
	"专区": {
		regexp.MustCompile(".+?专区$"),
	},
}

// GetChannelGroupName 根据频道名称自动获取分组名称
func GetChannelGroupName(channelName string) string {
	// 自动识别频道的分类
	for groupName, groupRules := range chGroupRuleMap {
		for _, groupRule := range groupRules {
			if groupRule.MatchString(channelName) {
				return groupName
			}
		}
	}
	return otherGroupName
}
