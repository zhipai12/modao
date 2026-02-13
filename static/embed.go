package static

import (
	"embed"
	"io/fs"
	"net/http"

	"github.com/gin-gonic/gin"
)

// 嵌入整个 dist 目录
//
//go:embed dist
var embeddedDistFS embed.FS

// SetupSPA 配置静态资源服务
func SetupSPA(engine *gin.Engine) {
	distFS, err := fs.Sub(embeddedDistFS, "dist")
	if err != nil {
		panic("FATAL: 无法加载嵌入资源！请确认 static/dist 目录存在: " + err.Error())
	}

	engine.StaticFS("/noProds", http.FS(distFS))

	engine.GET("/", func(c *gin.Context) {
		c.Redirect(http.StatusFound, "/noProds/")
	})
}
