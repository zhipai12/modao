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

// 可视化生成 mod和 dao配置
func noProdPages(engine *gin.Engine) {
    apiGroup := engine.Group("/noProd")
    apiGroup.GET("/genModDao/options", GenModDaoOptions()) // 获取选项
    apiGroup.POST("/genModDao/convert", GenModDaoConvert()) // 生成 mod dao 模块

    // 静态资源最后注册（/noProd 节点下仅有通配符，绝对安全）
    static.SetupSPA(engine)
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