package router

import (
	"context"
	"iptv/internal/app/config"
	"iptv/internal/app/iptv"
	"iptv/internal/app/iptv/hwctc"
	"iptv/internal/pkg/util"
	"net/http"
	"path"
	"strconv"
	"strings"
	"time"

	ginzap "github.com/gin-contrib/zap"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

var (
	logger *zap.Logger

	udpxyURLs      map[string]string
	catchupSources map[string]string
)

func NewEngine(ctx context.Context, conf *config.Config, interval time.Duration, udpxyURLCfg string) (*gin.Engine, error) {
	// L()：获取全局logger
	logger = zap.L()

	gin.SetMode(gin.ReleaseMode)

	// 获取程序运行的当前路径
	currDir, err := util.GetCurrentAbPathByExecutable()
	if err != nil {
		return nil, err
	}

	// 创建IPTV客户端
	iptvClient, err := newIPTVClient(conf)
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

	// 缓存udpxy配置
	udpxyURLs = parseUdpxyURLs(udpxyURLCfg)

	// 缓存回看请求参数配置
	catchupSources = conf.Catchup.Sources

	// 创建 Gin 路由引擎
	r := gin.New()

	// 日志记录
	r.Use(ginzap.Ginzap(logger, "", false))
	r.Use(ginzap.RecoveryWithZap(logger, true))

	// 查询直播源-m3u格式
	r.GET("/channel/m3u", GetM3UData)
	// 查询直播源-txt格式
	r.GET("/channel/txt", GetTXTData)
	// 查询直播源-pls格式
	r.GET("/channel/pls", GetPLSData)

	// 查询EPG-json格式
	r.GET("/epg/json", GetJsonEPG)
	// 查询EPG-xml格式
	r.GET("/epg/xml", GetXmlEPG)
	r.GET("/epg/xml.gz", GetXmlEPGWithGzip)

	// 查询频道logo
	r.Static("/logo", path.Join(currDir, "logos"))

	// 查询直播配置接口
	r.GET("/config/lives", GetLivesConfig)

	return r, nil
}

// parseUdpxyURLs 解析多个udpxy的URL
func parseUdpxyURLs(udpxyURLCfg string) map[string]string {
	result := make(map[string]string)

	if udpxyURLCfg == "" {
		return result
	}

	// 解析多个udpxy地址
	tmpUdpxyURLs := strings.Split(udpxyURLCfg, ",")
	for i, tmpUdpxyURL := range tmpUdpxyURLs {
		// 获取每个udpxy的名称和URL
		udpxyNameAndURL := strings.Split(tmpUdpxyURL, "=")
		if len(udpxyNameAndURL) != 2 {
			// 找不到名称则用序号代替
			result[strconv.Itoa(i)] = tmpUdpxyURL
		} else {
			result[udpxyNameAndURL[0]] = udpxyNameAndURL[1]
		}
	}
	return result
}

// initData 初始化数据
func initData(ctx context.Context, iptvClient iptv.Client) error {
	// 更新频道列表数据
	if err := updateChannelsWithRetry(ctx, iptvClient, 3); err != nil {
		return err
	}

	// 更新节目单
	if err := updateEPG(ctx, iptvClient); err != nil {
		logger.Error("Failed to update EPG.", zap.Error(err))
	}
	return nil
}

// newIPTVClient 读取配置文件并创建IPTV客户端
func newIPTVClient(conf *config.Config) (iptv.Client, error) {
	// 校验配置文件
	if err := conf.Validate(); err != nil {
		return nil, err
	}

	// 创建IPTV客户端
	return hwctc.NewClient(&http.Client{
		Timeout: 10 * time.Second,
	}, conf.HWCTC, conf.Key, conf.ServerHost, conf.Headers,
		conf.ChExcludeRule, conf.ChGroupRulesList, conf.ChLogoRuleList)
}
