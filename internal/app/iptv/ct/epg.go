package ct

import (
	"context"
	"errors"
	"iptv/internal/app/iptv"
)

var ErrParseChProgList = errors.New("failed to parse channel program list")
var ErrChProgListIsEmpty = errors.New("the list of programs is empty")

const (
	chProgAPILiveplay   = "liveplay_30"
	chProgAPIGdhdpublic = "gdhdpublic"
)

// GetAllChannelProgramList 获取所有频道的节目单列表
func (c *Client) GetAllChannelProgramList(ctx context.Context, channels []iptv.Channel) ([]iptv.ChannelProgramList, error) {
	// 请求认证的Token
	token, err := c.requestToken(ctx)
	if err != nil {
		return nil, err
	}

	epg := make([]iptv.ChannelProgramList, 0, len(channels))
	for _, channel := range channels {
		// 跳过不支持回看的频道
		if channel.TimeShift != "1" || channel.TimeShiftLength <= 0 {
			continue
		}

		var progList *iptv.ChannelProgramList
		switch c.config.ChannelProgramAPI {
		case chProgAPILiveplay:
			progList, err = c.getLiveplayChannelProgramList(ctx, token, &channel)
		case chProgAPIGdhdpublic:
			progList, err = c.getGdhdpublicChannelProgramList(ctx, token, &channel)
		default:
			progList, err = c.getLiveplayChannelProgramList(ctx, token, &channel)
		}

		if err != nil {
			c.logger.Sugar().Warnf("Failed to get the program list for channel %s. Error: %v", channel.ChannelName, err)
			continue
		}

		epg = append(epg, *progList)
	}

	return epg, nil
}
