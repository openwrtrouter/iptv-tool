package hwctc

import (
	"fmt"
	"iptv/internal/app/iptv"
	"net/http"

	"go.uber.org/zap"
)

type Client struct {
	httpClient *http.Client      // HTTP客户端
	config     *Config           // hwctc相关配置
	key        string            // 加密Authenticator的秘钥
	originHost string            // HTTP请求的服务器地址端口
	headers    map[string]string // 自定义HTTP请求头

	host string // 缓存最新重定向的服务器地址和端口

	logger *zap.Logger // 日志
}

var _ iptv.Client = (*Client)(nil)

func NewClient(httpClient *http.Client, config *Config, key, serverHost string, headers map[string]string) (iptv.Client, error) {
	// config不能为空
	if config == nil {
		return nil, fmt.Errorf("client config is nil")
	} else if err := config.Validate(); err != nil { // 校验config配置
		return nil, err
	}

	// 密钥和服务器地址必须配置
	if key == "" {
		return nil, fmt.Errorf("key is empty")
	} else if serverHost == "" {
		return nil, fmt.Errorf("serverHost is empty")
	}

	i := Client{
		httpClient: httpClient,
		config:     config,
		key:        key,
		originHost: serverHost,
		headers:    headers,
		host:       serverHost,
		logger:     zap.L(),
	}
	if i.httpClient == nil {
		i.httpClient = http.DefaultClient
	}
	return &i, nil
}

func (c *Client) setCommonHeaders(req *http.Request) {
	req.Header.Set("Host", c.host)
	// 设置自定义HTTP请求头
	if len(c.headers) > 0 {
		for k, v := range c.headers {
			req.Header.Set(k, v)
		}
	}
}
