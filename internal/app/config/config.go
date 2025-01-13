package config

import (
	"errors"
	"iptv/internal/app/iptv"
	"iptv/internal/app/iptv/hwctc"
	"os"
	"regexp"

	"go.uber.org/zap"
	"gopkg.in/yaml.v3"
)

type OptionChannelGroupRules struct {
	Name  string   `json:"name" yaml:"name"`   // 分组名称
	Rules []string `json:"rules" yaml:"rules"` // 分组规则
}

type Config struct {
	Key        string            `json:"key" yaml:"key"`               // 必填，8位数字，生成Authenticator的秘钥
	ServerHost string            `json:"serverHost" yaml:"serverHost"` // 必填，HTTP请求的IPTV服务器地址端口
	Headers    map[string]string `json:"headers" yaml:"headers"`       // 自定义HTTP请求头

	OptionChGroupRulesList []OptionChannelGroupRules `json:"chGroupRules" yaml:"chGroupRules"`
	ChGroupRulesList       []iptv.ChannelGroupRules  `json:"-" yaml:"-"` // Validate()时进行填充

	HWCTC *hwctc.Config `json:"hwctc,omitempty" yaml:"hwctc,omitempty"` // hw平台相关设置
}

func (c *Config) Validate() error {
	// 校验config配置
	if c.Key == "" ||
		c.ServerHost == "" {
		return errors.New("invalid IPTV-Tool config")
	}

	// L()：获取全局logger
	logger := zap.L()

	// 填充频道分组的正则表达式规则
	c.ChGroupRulesList = make([]iptv.ChannelGroupRules, 0, len(c.OptionChGroupRulesList))
	for _, opChGroupRules := range c.OptionChGroupRulesList {
		if opChGroupRules.Name == "" {
			logger.Warn("The channel group name is empty. Skip it.")
			continue
		} else if len(opChGroupRules.Rules) == 0 {
			logger.Warn("The channel group rule is empty. Skip it.", zap.String("groupName", opChGroupRules.Name))
			continue
		}

		rules := make([]*regexp.Regexp, 0, len(opChGroupRules.Rules))
		for _, ruleStr := range opChGroupRules.Rules {
			rule, err := regexp.Compile(ruleStr)
			if err != nil {
				logger.Warn("The channel group rule is incorrect. Skip it.", zap.String("groupName", opChGroupRules.Name), zap.String("rule", ruleStr), zap.Error(err))
				continue
			}

			rules = append(rules, rule)
		}
		if len(rules) > 0 {
			c.ChGroupRulesList = append(c.ChGroupRulesList, iptv.ChannelGroupRules{
				Name:  opChGroupRules.Name,
				Rules: rules,
			})
		}
	}

	return nil
}

func Load(fPath string) (*Config, error) {
	// 读取配置文件
	data, err := os.ReadFile(fPath)
	if err != nil {
		return nil, err
	}
	var config Config
	if err = yaml.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	return &config, nil
}

func CreateDefaultCfg(fPath string) error {
	// 写入默认配置
	f, err := os.Create(fPath)
	if err != nil {
		return err
	}
	defer f.Close()

	// 创建编码器
	encoder := yaml.NewEncoder(f)

	// 缺省配置
	defaultCfg := Config{
		ServerHost: "127.0.0.1",
		Headers: map[string]string{
			"Accept":           "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8",
			"User-Agent":       "Mozilla/5.0 (X11; Linux x86_64; Fhbw2.0) AppleWebKit",
			"Accept-Language":  "zh-CN,en-US;q=0.8",
			"X-Requested-With": "com.fiberhome.iptv",
		},
		OptionChGroupRulesList: []OptionChannelGroupRules{
			{
				Name: "央视",
				Rules: []string{
					"^(CCTV|中央).+?$",
				},
			},
			{
				Name: "卫视",
				Rules: []string{
					"^[^(热门)].+?卫视.*?$",
				},
			},
			{
				Name: "国际",
				Rules: []string{
					"^(CGTN|凤凰).+?$",
				},
			},
			{
				Name: "地方",
				Rules: []string{
					"^(SCTV|CDTV).+?$",
					"^(浙江|杭州|民生|钱江|教科影视|好易购|西湖|青少体育).+?$",
					"^(湖北|武汉).+?$",
				},
			},
			{
				Name: "专区",
				Rules: []string{
					".+?专区$",
				},
			},
		},
		HWCTC: &hwctc.Config{},
	}

	return encoder.Encode(&defaultCfg)
}
