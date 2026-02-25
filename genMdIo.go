package modao

import (
	"github.com/rrzu/cst"
	"github.com/rrzu/modao/common"
)

type (
	// OptionsRsp 获取选项 rsp
	OptionsRsp struct {
		DPT cst.Options[string, any] `json:"dpt"` // database:pattern:table 或 database:table
	}
)

// 生成 mod req rsp
type (
	// GenModDaoReq 生成 mod req
	GenModDaoReq struct {
		DBType      common.ConnectType `form:"dbType" json:"dbType" binding:"required"`           // 数据库类型
		ConnectName common.ConnectName `form:"connectName" json:"connectName" binding:"required"` // 连接名称
		DBName      string             `form:"dbName" json:"dbName" binding:"required"`           // 数据库名称
		PatternName string             `form:"patternName" json:"patternName"`                    // 数据库模式名称
		TableName   string             `form:"tableName" json:"tableName" binding:"required"`     // 表名
		IsCover     bool               `form:"isCover" json:"isCover"`                            // 是否覆盖
	}

	// GenModDaoRsp 生成 mod rsp
	GenModDaoRsp struct {
		Imports string `json:"imports"` // imports 列表
		Mod     string `json:"mod"`     // mod 文本
		Dao     string `json:"dao"`     // dao 文本
	}
)

// NamesToOptions []string 转成 constMp.Options
func NamesToOptions[T string, S any](d []string, typ cst.DataType) cst.Options[T, S] {
	var options = cst.Options[T, S]{
		Typ:  typ,
		Opts: make([]cst.Option[T, S], len(d)),
	}
	for i, name := range d {
		options.Opts[i].CnName = name
		options.Opts[i].Val = name
	}
	return options
}
