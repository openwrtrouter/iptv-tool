package iptv

import (
	"context"
)

type Client interface {
	// GetAllChannelList 获取频道列表
	GetAllChannelList(ctx context.Context) ([]Channel, error)

	// GetAllChannelProgramList 获取所有频道的节目单列表
	GetAllChannelProgramList(ctx context.Context, channels []Channel) ([]ChannelProgramList, error)
}
