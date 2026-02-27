package modao

import (
	"fmt"

	"github.com/rrzu/cst"
)

// GenModDaoOptions 获取选项
func GenModDaoOptions() (rsp OptionsRsp, err error) {
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
			CnName: string(connectData.ConnectInfo.ConnectName),
			Val:    string(connectData.ConnectInfo.ConnectType),
			Sub:    &opts,
		})
	}

	return
}

// GenModDaoConvert 生成 mod dao 模块
func GenModDaoConvert(req GenModDaoReq) (rsp GenModDaoRsp, err error) {
	// 生成常量
	err = GenConstContent(req)
	if err != nil {
		err = fmt.Errorf("生成常量错误%v", err)
		return
	}

	// 生成文件
	filename, err := GenMdFile(req)
	if err != nil {
		err = fmt.Errorf("生成文件错误%v", err)
		return
	}

	// 获取分析语法树对象
	mng, err := GetModStmt(req)
	if err != nil {
		err = fmt.Errorf("获取分析语法树对象错误%v", err)
		return
	}

	// 建表语句
	createTbl := mng.CreateTbl()

	// 解析建表语句
	table, err := mng.GenerateModTable(createTbl)
	if err != nil {
		err = fmt.Errorf("解析建表语句错误%v", err)
		return
	}

	// 生成 GenImport GenModCols、GenDaoCols 结构
	var imports = table.ToGenImport(req)
	var daoCols = table.ToGenDaoCols(req)
	var modCols = table.ToGenModCols(req)

	// 生成 import 部分
	err = GenAppendToFile(filename, imports.ToText())
	if err != nil {
		err = fmt.Errorf("添加 import 部分错误%v", err)
		return
	}

	// 生成 dao 部分
	err = GenAppendToFile(filename, daoCols.ToText())
	if err != nil {

		err = fmt.Errorf("添加 dao 部分错误%v", err)
		return
	}

	// 生成 mod 部分
	err = GenAppendToFile(filename, modCols.ToText())
	if err != nil {
		err = fmt.Errorf("添加 mod 部分错误%v", err)
		return
	}

	// 转成展示文本
	rsp.Imports = imports.ToText()
	rsp.Mod = modCols.ToText()
	rsp.Dao = daoCols.ToText()

	return
}
