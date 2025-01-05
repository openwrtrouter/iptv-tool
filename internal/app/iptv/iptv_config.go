package iptv

type Config struct {
	Key           string `json:"key"`           // 8位数字，加密Authenticator的秘钥，每个机顶盒可能都不同，获取频道列表必须使用
	InterfaceName string `json:"interfaceName"` // 网络接口的名称。若配置则生成Authenticator时，优先使用该接口对应的IPv4地址，而不使用`ip`字段的值。
	// 以下信息均可通过抓包获取
	ServerHost string `json:"serverHost"` // HTTP请求的服务器地址端口
	IP         string `json:"ip"`         // 生成Authenticator所需的IP地址。可随便一个地址，或者通过配置`interfaceName`动态获取
}
