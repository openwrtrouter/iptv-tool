package hwctc

import (
	"errors"
	"iptv/internal/app/iptv"
)

type Config struct {
	iptv.Config // 公共配置

	// 以下信息均可通过抓包获取
	XRequestedWith    string `json:"x-requested-with,omitempty"`  // HTTP请求时需要携带的请求头，找不到可不填写
	ChannelProgramAPI string `json:"channelProgramAPI,omitempty"` // 请求频道节目信息的API接口，目前只支持两种：liveplay_30或者gdhdpublic，缺省为liveplay_30。
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
		c.STBID == "" ||
		c.MAC == "" ||
		c.SoftwareVersion == "" {
		return errors.New("invalid IPTV Client config")
	}

	return nil
}
