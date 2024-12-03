package router

import (
	"context"
	"errors"
	"fmt"
	"iptv/internal/app/iptv"
	"net/http"
	"sync/atomic"
	"time"

	"github.com/gin-gonic/gin"
)

var (
	// 缓存最新的频道列表数据
	channelsPtr atomic.Pointer[[]iptv.Channel]
)

// GetM3UData 查询直播源m3u
func GetM3UData(c *gin.Context) {
	channels := *channelsPtr.Load()

	if len(channels) == 0 {
		c.Status(http.StatusNotFound)
		return
	}

	// 将获取到的频道列表转换为m3u格式
	m3uContent, err := iptv.ToM3UFormat(channels, udpxyURL)
	if err != nil {
		fmt.Println("Failed to convert channel list to m3u format: ", err)
		// 返回响应
		c.Status(http.StatusOK)
		return
	}

	// 返回响应
	c.String(http.StatusOK, m3uContent)
}

// GetTXTData 查询直播源txt
func GetTXTData(c *gin.Context) {
	channels := *channelsPtr.Load()

	if len(channels) == 0 {
		c.Status(http.StatusNotFound)
		return
	}

	// 将获取到的频道列表转换为txt格式
	txtContent, err := iptv.ToTxtFormat(channels, udpxyURL)
	if err != nil {
		fmt.Println("Failed to convert channel list to txt format: ", err)
		// 返回响应
		c.Status(http.StatusOK)
		return
	}

	// 返回响应
	c.String(http.StatusOK, txtContent)
}

// updateChannelsWithRetry 更新缓存的频道数据（失败重试）
func updateChannelsWithRetry(ctx context.Context, iptvClient *iptv.Client, maxRetries int) error {
	var err error
	for i := 0; i < maxRetries; i++ {
		if err = updateChannels(ctx, iptvClient); err != nil {
			fmt.Printf("Failed to update channel list, will try again after waiting %d seconds. Error: %v, number of retries: %d.\n", waitSeconds, err, i)
			time.Sleep(waitSeconds * time.Second)
		} else {
			break
		}
	}
	return err
}

// updateChannels 更新缓存的频道数据
func updateChannels(ctx context.Context, iptvClient *iptv.Client) error {
	// 登录认证获取Token等信息
	token, err := iptvClient.GenerateToken(ctx)
	if err != nil {
		return err
	}

	// 查询最新的频道列表
	channels, err := iptvClient.GetChannelList(ctx, token)
	if err != nil {
		return err
	}

	if len(channels) == 0 {
		return errors.New("no channels found")
	}

	fmt.Printf("The channel list has been updated, rows: %d.\n", len(channels))
	// 更新缓存的频道列表
	channelsPtr.Store(&channels)

	return nil
}
