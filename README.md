# Modao



## 要求
* 需要确保本地安装的 go 环境版本大于 1.23

## 安装
```bash
go get https://github.com/rrzu/modao
```

## 使用
基础配置
```go
import "github.com/rrzu/modao"

// 初始化 modao 配置
(&modao.Config{}).
SetDebugKey("onDebug").
SetGormDb(modao.ConnectInfo{ConnectName: _const.ConnectNameMysqlShenzhen, ConnectType: common.ConnectTypeMysql}, db.Clickhouse(false)).
Init()
```

可视化生成mod和dao配置
```go
import (
    "github.com/rrzu/modao"
    "github.com/rrzu/modao/static"
)

// 初始化 modao 配置
(&modao.Config{}).
SetDebugKey("onDebug").
SetGormDb(modao.ConnectInfo{ConnectName: _const.ConnectNameMysqlShenzhen, ConnectType: common.ConnectTypeMysql}, db.Clickhouse(false)).
SetGenMdPath(filepath.Join((func() string { wd, _ := os.Getwd(); return wd })(), "repository")). // 生成 mod 和 dao 路径
Init()

// 引入静态文件以及配置路由
func NoProdPages(engine *gin.Engine) {
    staticHandler, _ := static.StaticHandler()

    // 将 /noProds 重定向到 /noProds/
    engine.GET(static.StaticPath, func(c *gin.Context) {
        c.Redirect(http.StatusMovedPermanently, static.StaticPath+"/")
    })

    // 处理所有 /noProds/ 下的子路径
    engine.Any(static.StaticPath+"/*filepath", gin.WrapH(staticHandler))

    // 生成 mod 和 dao 路由
    var noProdGenModDaoGroupRoute = engine.Group("/noProd/genModDao")
    noProdGenModDaoGroupRoute.GET("/options", GenModDaoOptions)
    noProdGenModDaoGroupRoute.POST("/convert", GenModDaoConvert)
}

func GenModDaoOptions(ctx *gin.Context) {
    rsp, err := modao.GenModDaoOptions()
    if err != nil {
        ctx.JSON(http.StatusOK, gin.H{
        "code":    500,
        "message": err.Error(),
    })
		return
    }

    ctx.JSON(http.StatusOK, gin.H{
        "code":    200,       // 业务状态码（0=成功，按项目规范调整）
        "message": "success", // 提示信息
        "data":    rsp,       // 原始响应数据
    })
	
	return
}

func GenModDaoConvert(ctx *gin.Context) {
    req := modao.GenModDaoReq{}
    if err := ctx.ShouldBindJSON(&req); err != nil {
        ctx.JSON(http.StatusOK, gin.H{
        "code":    500,
        "message": err.Error(),
        })
        return
    }

    rsp, err := modao.GenModDaoConvert(req)
    if err != nil {
        ctx.JSON(http.StatusOK, gin.H{
        "code":    500,
        "message": err.Error(),
        })
        return
    }

    ctx.JSON(http.StatusOK, gin.H{
        "code":    200,       // 业务状态码（0=成功，按项目规范调整）
        "message": "success", // 提示信息
        "data":    rsp,       // 原始响应数据
    })
	return
}


```

可视化访问地址:
```shell
http://localhost/noProds/#/home
```