package iptv

import (
	"errors"
)

type Config struct {
	Key           string `json:"key"`           // 8位数字，加密Authenticator的秘钥，每个机顶盒可能都不同，获取频道列表必须使用
	InterfaceName string `json:"interfaceName"` // 网络接口的名称。若配置则生成Authenticator时，优先使用该接口对应的IPv4地址，而不使用`ip`字段的值。
	// 以下信息均可通过抓包获取
	ServerHost string `json:"serverHost"` // HTTP请求的服务器地址端口
	IP         string `json:"ip"`         // 生成Authenticator所需的IP地址。可随便一个地址，或者通过配置`interfaceName`动态获取
	// HTTP请求时需要携带的请求头，找不到可不填写
	XRequestedWith string `json:"x-requested-with,omitempty"`
	// 以下信息均可通过抓包请求ValidAuthenticationHWCTC.jsp的参数拿到
	UserID           string `json:"userID"`
	Lang             string `json:"lang,omitempty"`      // 如果没有可以不填
	NetUserID        string `json:"netUserId,omitempty"` // 如果没有可以不填
	STBType          string `json:"stbType"`
	STBVersion       string `json:"stbVersion"`
	Conntype         string `json:"conntype"`
	STBID            string `json:"stbID"` // 机顶盒背面也可查
	TemplateName     string `json:"templateName"`
	AreaId           string `json:"areaId"`
	UserGroupId      string `json:"userGroupId,omitempty"`
	ProductPackageId string `json:"productPackageId,omitempty"`
	MAC              string `json:"mac"` // 机顶盒背面也可查
	UserField        string `json:"userField,omitempty"`
	SoftwareVersion  string `json:"softwareVersion"`
	IsSmartStb       string `json:"isSmartStb,omitempty"`
	Vip              string `json:"vip,omitempty"`
}

func (c *Config) Validate() error {
	// 校验config配置
	if c.Key == "" ||
		c.ServerHost == "" ||
		(c.IP == "" && c.InterfaceName == "") ||
		c.UserID == "" ||
		c.STBType == "" ||
		c.STBVersion == "" ||
		c.Conntype == "" ||
		c.STBID == "" ||
		c.TemplateName == "" ||
		c.MAC == "" ||
		c.SoftwareVersion == "" {
		return errors.New("invalid IPTV Client config")
	}

	return nil
}
