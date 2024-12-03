package router

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type Lives struct {
	Lives []Live `json:"lives"`
}

// Live 直播配置
type Live struct {
	Name string `json:"name"`
	Boot bool   `json:"boot"`
	Url  string `json:"url"`
	Epg  string `json:"epg,omitempty"`
	Logo string `json:"logo,omitempty"`
}

var lives *Lives

// LoadLivesConfig 加载直播配置
func LoadLivesConfig(livesCfg *Lives) {
	lives = livesCfg
}

// GetLivesConfig 查询直播配置
func GetLivesConfig(c *gin.Context) {
	if lives == nil {
		c.Status(http.StatusNotFound)
		return
	}

	// 返回响应
	c.JSON(http.StatusOK, lives)
}
