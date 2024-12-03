package router

import (
	"context"
	"iptv/internal/app/iptv"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

var udpxyURL string

func NewEngine(ctx context.Context, interval time.Duration, udpxyURLCfg string) (*gin.Engine, error) {
	gin.SetMode(gin.ReleaseMode)

	// 缓存udpxy配置
	udpxyURL = udpxyURLCfg

	// 创建IPTV客户端
	iptvClient, err := newIPTVClient()
	if err != nil {
		return nil, err
	}

	// 执行初始化操作
	err = initData(ctx, iptvClient)
	if err != nil {
		return nil, err
	}

	// 执行定时任务
	Schedule(ctx, iptvClient, interval)

	// 创建 Gin 路由引擎
	r := gin.Default()

	// 查询直播源-m3u格式
	r.GET("/channel/m3u", GetM3UData)
	// 查询直播源-txt格式
	r.GET("/channel/txt", GetTXTData)

	// 查询EPG-json格式
	r.GET("/epg/json", GetJsonEPG)
	// 查询EPG-xml格式
	r.GET("/epg/xml", GetXmlEPG)

	// 查询直播配置接口
	r.GET("/config/lives", GetLivesConfig)

	return r, nil
}

// initData 初始化数据
func initData(ctx context.Context, iptvClient *iptv.Client) error {
	// 更新频道列表数据
	if err := updateChannelsWithRetry(ctx, iptvClient, 3); err != nil {
		return err
	}

	// 更新节目单
	if err := updateEPG(ctx, iptvClient); err != nil {
		return err
	}
	return nil
}

// newIPTVClient 读取配置文件并创建IPTV客户端
func newIPTVClient() (*iptv.Client, error) {
	// 读取IPTV配置
	var config iptv.Config
	err := viper.Unmarshal(&config)
	if err != nil {
		return nil, err
	}

	// 创建IPTV客户端
	return iptv.NewClient(&http.Client{
		Timeout: 10 * time.Second,
	}, &config)
}
