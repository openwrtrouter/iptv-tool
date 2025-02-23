package router

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type Lives struct {
	Lives []Live `json:"lives"`
}

// Live 直播配置
type Live map[string]any

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
	c.PureJSON(http.StatusOK, lives)
}
