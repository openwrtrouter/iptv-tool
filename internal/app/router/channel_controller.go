package router

import (
	"context"
	"errors"
	"fmt"
	"iptv/internal/app/iptv"
	"net/http"
	"strconv"
	"sync/atomic"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

const (
	diypCatchupSource string = "?playseek=${(b)yyyyMMddHHmmss}-${(e)yyyyMMddHHmmss}"
	kodiCatchupSource string = "?playseek={utc:YmdHMS}-{utcend:YmdHMS}"
)

var (
	// 缓存最新的频道列表数据
	channelsPtr atomic.Pointer[[]iptv.Channel]
)

// GetM3UData 查询直播源m3u
func GetM3UData(c *gin.Context) {
	// 获取catchup-source格式
	var catchupSource string
	csFormat := c.DefaultQuery("csFormat", "0")
	switch csFormat {
	case "1":
		catchupSource = kodiCatchupSource
	default:
		catchupSource = diypCatchupSource
	}

	// 是否优先是由组播地址
	multiFirstStr := c.DefaultQuery("multiFirst", "true")
	multicastFirst, err := strconv.ParseBool(multiFirstStr)
	if err != nil {
		multicastFirst = true
	}

	channels := *channelsPtr.Load()
	if len(channels) == 0 {
		c.Status(http.StatusNotFound)
		return
	}

	// 设置台标的统一Base URL
	logoBaseUrl := fmt.Sprintf("http://%s/logo", c.Request.Host)

	// 将获取到的频道列表转换为m3u格式
	m3uContent, err := iptv.ToM3UFormat(channels, udpxyURL, catchupSource, multicastFirst, logoBaseUrl)
	if err != nil {
		logger.Error("Failed to convert channel list to m3u format.", zap.Error(err))
		// 返回响应
		c.Status(http.StatusOK)
		return
	}

	// 返回响应
	c.String(http.StatusOK, m3uContent)
}

// GetTXTData 查询直播源txt
func GetTXTData(c *gin.Context) {
	// 是否优先是由组播地址
	multiFirstStr := c.DefaultQuery("multiFirst", "true")
	multicastFirst, err := strconv.ParseBool(multiFirstStr)
	if err != nil {
		multicastFirst = true
	}

	channels := *channelsPtr.Load()
	if len(channels) == 0 {
		c.Status(http.StatusNotFound)
		return
	}

	// 将获取到的频道列表转换为txt格式
	txtContent, err := iptv.ToTxtFormat(channels, udpxyURL, multicastFirst)
	if err != nil {
		logger.Error("Failed to convert channel list to txt format.", zap.Error(err))
		// 返回响应
		c.Status(http.StatusOK)
		return
	}

	// 返回响应
	c.String(http.StatusOK, txtContent)
}

// updateChannelsWithRetry 更新缓存的频道数据（失败重试）
func updateChannelsWithRetry(ctx context.Context, iptvClient iptv.Client, maxRetries int) error {
	var err error
	for i := 0; i < maxRetries; i++ {
		if err = updateChannels(ctx, iptvClient); err != nil {
			logger.Sugar().Errorf("Failed to update channel list, will try again after waiting %d seconds. Error: %v, number of retries: %d.", waitSeconds, err, i)
			time.Sleep(waitSeconds * time.Second)
		} else {
			break
		}
	}
	return err
}

// updateChannels 更新缓存的频道数据
func updateChannels(ctx context.Context, iptvClient iptv.Client) error {
	// 查询最新的频道列表
	channels, err := iptvClient.GetAllChannelList(ctx)
	if err != nil {
		return err
	}

	if len(channels) == 0 {
		return errors.New("no channels found")
	}

	logger.Sugar().Infof("The channel list has been updated, rows: %d.", len(channels))
	// 更新缓存的频道列表
	channelsPtr.Store(&channels)

	return nil
}
