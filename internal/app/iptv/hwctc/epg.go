package hwctc

import (
	"context"
	"errors"
	"iptv/internal/app/iptv"
	"slices"

	"go.uber.org/zap"
)

var (
	ErrParseChProgList   = errors.New("failed to parse channel program list")
	ErrChProgListIsEmpty = errors.New("the list of programs is empty")
	ErrEPGApiNotFound    = errors.New("epg api not found")
)

const (
	maxBackDay = 8

	chProgAPILiveplay        = "liveplay_30"
	chProgAPIGdhdpublic      = "gdhdpublic"
	chProgAPIVsp             = "vsp"
	chProgAPIStbEpg2023Group = "StbEpg2023Group"
	chProgAPIDefaulttrans2   = "defaulttrans2"
)

type getChannelProgramListFunc func(ctx context.Context, token *Token, channel *iptv.Channel) (*iptv.ChannelProgramList, error)

// GetAllChannelProgramList 获取所有频道的节目单列表
func (c *Client) GetAllChannelProgramList(ctx context.Context, channels []iptv.Channel) ([]iptv.ChannelProgramList, error) {
	// 请求认证的Token
	token, err := c.requestToken(ctx)
	if err != nil {
		return nil, err
	}

	var result []iptv.ChannelProgramList
	switch c.config.ChannelProgramAPI {
	case chProgAPILiveplay:
		result, err = c.getAllChannelProgramList(ctx, channels, token, c.getLiveplayChannelProgramList)
	case chProgAPIGdhdpublic:
		result, err = c.getAllChannelProgramList(ctx, channels, token, c.getGdhdpublicChannelProgramList)
	case chProgAPIVsp:
		result, err = c.getAllChannelProgramList(ctx, channels, token, c.getVspChannelProgramList)
	case chProgAPIStbEpg2023Group:
		result, err = c.getStbEpg2023GroupAllChannelProgramList(ctx, channels, token)
	case chProgAPIDefaulttrans2:
		result, err = c.getAllChannelProgramList(ctx, channels, token, c.getDefaulttrans2ChannelProgramList)
	default:
		// 自动选择调用EPG的API接口
		result, err = c.getAllChannelProgramListByAuto(ctx, channels, token)
	}

	return result, err
}

// getAllChannelProgramList 获取所有频道的节目单列表
func (c *Client) getAllChannelProgramList(ctx context.Context, channels []iptv.Channel, token *Token, getChProgFunc getChannelProgramListFunc) ([]iptv.ChannelProgramList, error) {
	epg := make([]iptv.ChannelProgramList, 0, len(channels))
	for _, channel := range channels {
		// 跳过不支持回看的频道
		if channel.TimeShift != "1" || channel.TimeShiftLength <= 0 {
			continue
		}

		progList, err := getChProgFunc(ctx, token, &channel)
		if err != nil {
			if errors.Is(err, ErrEPGApiNotFound) {
				return nil, err
			}
			c.logger.Sugar().Warnf("Failed to get the program list for channel %s. Error: %v", channel.ChannelName, err)
			continue
		}

		if progList != nil && len(progList.DateProgramList) > 0 {
			// 对频道的节目单按日期升序排序
			slices.SortFunc(progList.DateProgramList, func(a, b iptv.DateProgram) int {
				return a.Date.Compare(b.Date)
			})

			epg = append(epg, *progList)
		}
	}
	return epg, nil
}

// getAllChannelProgramListByAuto 自动选择调用EPG的API接口
func (c *Client) getAllChannelProgramListByAuto(ctx context.Context, channels []iptv.Channel, token *Token) ([]iptv.ChannelProgramList, error) {
	result, err := c.getAllChannelProgramList(ctx, channels, token, c.getLiveplayChannelProgramList)
	if !errors.Is(err, ErrEPGApiNotFound) {
		c.logger.Info("An available EPG API was found.", zap.String("channelProgramAPI", chProgAPILiveplay))
		c.config.ChannelProgramAPI = chProgAPILiveplay
		return result, err
	}

	result, err = c.getAllChannelProgramList(ctx, channels, token, c.getGdhdpublicChannelProgramList)
	if !errors.Is(err, ErrEPGApiNotFound) {
		c.logger.Info("An available EPG API was found.", zap.String("channelProgramAPI", chProgAPIGdhdpublic))
		c.config.ChannelProgramAPI = chProgAPIGdhdpublic
		return result, err
	}

	result, err = c.getAllChannelProgramList(ctx, channels, token, c.getVspChannelProgramList)
	if !errors.Is(err, ErrEPGApiNotFound) {
		c.logger.Info("An available EPG API was found.", zap.String("channelProgramAPI", chProgAPIVsp))
		c.config.ChannelProgramAPI = chProgAPIVsp
		return result, err
	}

	result, err = c.getStbEpg2023GroupAllChannelProgramList(ctx, channels, token)
	if !errors.Is(err, ErrEPGApiNotFound) {
		c.logger.Info("An available EPG API was found.", zap.String("channelProgramAPI", chProgAPIStbEpg2023Group))
		c.config.ChannelProgramAPI = chProgAPIStbEpg2023Group
		return result, err
	}

	result, err = c.getAllChannelProgramList(ctx, channels, token, c.getDefaulttrans2ChannelProgramList)
	if !errors.Is(err, ErrEPGApiNotFound) {
		c.logger.Info("An available EPG API was found.", zap.String("channelProgramAPI", chProgAPIDefaulttrans2))
		c.config.ChannelProgramAPI = chProgAPIDefaulttrans2
		return result, err
	}

	c.logger.Warn("No suitable EPG API found.")
	return nil, err
}
