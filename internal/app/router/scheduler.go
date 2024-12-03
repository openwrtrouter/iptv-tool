package router

import (
	"context"
	"fmt"
	"iptv/internal/app/iptv"
	"time"
)

const waitSeconds = 30

// Schedule 定时调度更新缓存数据
func Schedule(ctx context.Context, iptvClient *iptv.Client, duration time.Duration) {
	// 创建定时任务
	ticker := time.NewTicker(duration)
	go func() {
		for {
			select {
			case <-ctx.Done():
				fmt.Println("The scheduling task has been stopped.")
				return
			case <-ticker.C:
				fmt.Println("Start executing the scheduling task.")

				// 更新频道列表数据
				if err := updateChannelsWithRetry(ctx, iptvClient, 3); err != nil {
					fmt.Println("Failed to update channel list:", err)
				}

				// 更新节目单数据
				if err := updateEPG(ctx, iptvClient); err != nil {
					fmt.Println("Failed to update EPG:", err)
				}

				fmt.Println("The scheduling task has been completed.")
			}
		}
	}()
}
