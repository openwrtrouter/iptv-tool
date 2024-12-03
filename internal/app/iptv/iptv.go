package iptv

import (
	"net/http"
)

type Client struct {
	httpClient *http.Client // HTTP客户端
	config     *Config      // IPTV配置
	host       string       // 缓存最新重定向的服务器地址和端口
}

func NewClient(httpClient *http.Client, config *Config) (*Client, error) {
	// 校验config配置
	if err := config.Validate(); err != nil {
		return nil, err
	}

	i := Client{
		httpClient: httpClient,
		host:       config.ServerHost,
		config:     config,
	}
	if i.httpClient == nil {
		i.httpClient = http.DefaultClient
	}
	return &i, nil
}

func (c *Client) setCommonHeaders(req *http.Request) {
	req.Header.Set("Host", c.host)
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8")
	req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64; Fhbw2.0) AppleWebKit/534.24 (KHTML, like Gecko) Safari/534.24 chromium/webkit")
	req.Header.Set("Accept-Language", "zh-CN,en-US;q=0.8")
	if c.config.XRequestedWith != "" {
		req.Header.Set("X-Requested-With", c.config.XRequestedWith)
	} else {
		req.Header.Set("X-Requested-With", "com.fiberhome.iptv")
	}
}
