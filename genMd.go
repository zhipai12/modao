package modao

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rrzu/cst"
	"github.com/rrzu/modao/common"
)

// noProdPages 引入示例
// func noProdPages(engine *gin.Engine) {
//	apiGroup := engine.Group("/noProd")
//	apiGroup.GET("/genModDao/options", GenModDaoOptions)  // 获取选项
//	apiGroup.POST("/genModDao/convert", GenModDaoConvert) // 生成 mod dao 模块
//
//	静态资源最后注册（/noProd 节点下仅有通配符，绝对安全）
//	static.SetupSPA(engine)
// }

// GenModDaoOptions 获取选项
func GenModDaoOptions(ctx *gin.Context) {
	var rsp OptionsRsp

	rsp.DPT = cst.Options[string, any]{
		Typ:  cst.DataTypeString,
		Opts: make([]cst.Option[string, any], 0),
	}

	for _, connectData := range GetAllConnectData() {
		opts, err := Options(connectData)
		if err != nil {
			continue
		}

		rsp.DPT.Opts = append(rsp.DPT.Opts, cst.Option[string, any]{
			CnName: string(connectData.ConnectType),
			Val:    string(connectData.ConnectType),
			Sub:    &opts,
		})

	}

	ctx.JSON(http.StatusOK, common.StdSuccess(rsp, "成功"))

	return
}

// GenModDaoConvert 生成 mod dao 模块
func GenModDaoConvert(ctx *gin.Context) {
	req := GenModDaoReq{}
	if err := ctx.ShouldBind(&req); err != nil {
		ctx.JSON(http.StatusOK, common.StdFail("参数错误", 500))
		return
	}

	filename, err := GenMdFile(req)
	if err != nil {
		ctx.JSON(http.StatusOK, common.StdFail("生成文件错误", 500))
		return
	}

	// filename := "app\\repository\\ms\\communication\\md\\ChatWinEffected_gen.go"

	// 获取分析语法树对象
	mng, err := GetModStmt(req)
	if err != nil {
		ctx.JSON(http.StatusOK, common.StdFail("获取分析语法树对象错误", 500))
		return
	}

	// 建表语句
	createTbl := mng.CreateTbl()

	// 解析建表语句
	table, err := mng.GenerateModTable(createTbl)
	if err != nil {
		ctx.JSON(http.StatusOK, common.StdFail("解析建表语句错误", 500))
		return
	}

	// 生成 GenImport GenModCols、GenDaoCols 结构
	var imports = table.ToGenImport(req)
	var daoCols = table.ToGenDaoCols(req)
	var modCols = table.ToGenModCols(req)

	// 生成 import 部分
	err = GenAppendToFile(filename, imports.ToText())
	if err != nil {
		ctx.JSON(http.StatusOK, common.StdFail("添加 import 部分错误", 500))
		return
	}

	// 生成 dao 部分
	err = GenAppendToFile(filename, daoCols.ToText())
	if err != nil {
		ctx.JSON(http.StatusOK, common.StdFail("添加 dao 部分错误", 500))
		return
	}

	// 生成 mod 部分
	err = GenAppendToFile(filename, modCols.ToText())
	if err != nil {
		ctx.JSON(http.StatusOK, common.StdFail("添加 mod 部分错误", 500))
		return
	}

	// 转成展示文本
	var rsp GenModDaoRsp
	rsp.Imports = imports.ToText()
	rsp.Mod = modCols.ToText()
	rsp.Dao = daoCols.ToText()

	ctx.JSON(http.StatusOK, common.StdSuccess(rsp, "成功"))
	return
}
