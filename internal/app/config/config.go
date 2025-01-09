package config

import (
	"errors"
	"iptv/internal/app/iptv/hwctc"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Key        string            `json:"key" yaml:"key"`               // 必填，8位数字，生成Authenticator的秘钥
	ServerHost string            `json:"serverHost" yaml:"serverHost"` // 必填，HTTP请求的IPTV服务器地址端口
	Headers    map[string]string `json:"headers" yaml:"headers"`       // 自定义HTTP请求头

	HWCTC *hwctc.Config `json:"hwctc,omitempty" yaml:"hwctc,omitempty"` // hw平台相关设置
}

func (c *Config) Validate() error {
	// 校验config配置
	if c.Key == "" ||
		c.ServerHost == "" {
		return errors.New("invalid IPTV-Tool config")
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
