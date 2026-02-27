package static

import (
	"embed"
	"io/fs"
	"net/http"
)

//go:embed dist
var embeddedDistFS embed.FS

// StaticPath 静态资源路径前缀
const StaticPath = "/noProds"

// StaticHandler 返回处理静态资源的 Handler。
func StaticHandler() (http.Handler, error) {
	distFS, err := fs.Sub(embeddedDistFS, "dist")
	if err != nil {
		return nil, err
	}

	return http.StripPrefix(StaticPath, http.FileServer(http.FS(distFS))), nil
}
