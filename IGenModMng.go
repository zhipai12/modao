package modao

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/rrzu/cst"
	"github.com/rrzu/modao/common"
)

type (
	// IGenModMng 生成 mod 的 interface
	IGenModMng interface {
		Options() (rsp cst.Options[string, any], err error)             // 获取选项
		SetGenModReq(req GenModDaoReq)                                  // 设置 req
		CreateTbl() (dest TblFromDb)                                    // 获取建表语句
		GenerateModTable(createTbl TblFromDb) (tbl ModTable, err error) // 通过建表语句生成 ModTable
	}

	// ModTable 表
	ModTable struct {
		tableName    string    // 表名
		tableComment string    // 表注释
		columns      []Columns // 字段注释
	}

	// Columns 字段
	Columns struct {
		_index int // 索引

		field   string // 字段名
		typ     string // 字段类型
		tag     string // tag
		comment string // 注释
	}

	// TblFromDb 获取表
	TblFromDb struct {
		CreateTable string `gorm:"column:Create Table"`
	}
)

// Options 获取数据库列表
func Options(connect ConnectData) (options cst.Options[string, any], err error) {
	var mng IGenModMng
	if mng, err = getModStmt(connect); err != nil {
		return
	}

	if mng == nil {
		return
	}

	// 获取数据库列表
	return mng.Options()
}

func getModStmt(connect ConnectData) (mng IGenModMng, err error) {
	switch connect.ConnectType {
	case common.ConnectTypeMysql:
		mng = &GenMysqlModMng{
			ConnectData: connect,
		}
	//case _const.DBTypeCh:
	//case _const.DBTypePg:
	//case _const.DBTypeHo:
	//	mng = &GenHologresModMng{}
	default:
		err = fmt.Errorf("错误类型")
	}
	return
}

// GenMdFile 生成 mod and dao 文件
func GenMdFile(req GenModDaoReq) (filename string, err error) {
	var partPath string

	switch req.DBType {
	case common.ConnectTypeMysql:
		partPath = filepath.Join("ms", req.DBName, "md")
	case common.ConnectTypeHologres:
		partPath = filepath.Join("hg", "data_analysis", req.PatternName, "md")

	default:
		err = errors.New("该数据库类型暂不支持，请重新选择数据库类型")
		return
	}

	// 构建文件路径
	filename = filepath.Join(OutputDir, partPath, common.UnderScoreToCamel(req.TableName)+"_gen.go")

	// 转为绝对路径（防御性编程）
	absDir, err := filepath.Abs(filename)
	if err != nil {
		err = errors.New("获取绝对路径失败")
	}

	// 检查文件是否存在且不允许覆盖
	if common.CheckFileIsExist(absDir) && !req.IsCover {
		err = errors.New("文件已存在，如需覆盖请选择覆盖")
		return
	}

	// 创建文件
	if err = common.CreateFile(absDir); err != nil {
		err = fmt.Errorf("创建文件失败: %w", err)
		return
	}

	return
}

// GetModStmt 根据数据库类型 获取分析语法树对象
func GetModStmt(req GenModDaoReq) (mng IGenModMng, err error) {
	var connect ConnectData
	connect.ConnectType = req.DBType

	if mng, err = getModStmt(connect); err != nil {
		return
	}

	mng.SetGenModReq(req)

	return
}

// toTag 表名转换成 tag
func toTag(fieldName string) string {
	if fieldName == "inv_created_at" || fieldName == "inv_updated_at" {
		return fmt.Sprintf("`json:\"%s\" gorm:\"column:%s;autoCreateTime\"`", fieldName, fieldName)
	}
	return fmt.Sprintf("`json:\"%s\" gorm:\"column:%s\"`", fieldName, fieldName)
}

// 获取 mod struct 名
func modName(tblName string) string {
	return common.UnderScoreToCamel(tblName) + "Mod"
}

// h获取 dao struct 名
func daoName(tblName string) string {
	return common.UnderScoreToCamel(tblName) + "Dao"
}

// 创建 mod struct 的方法
func createModMethods(req GenModDaoReq) (col []string) {
	col = append(col, fmt.Sprintf("\nfunc (m *%s) TableName() string {", modName(req.TableName)))
	col = append(col, fmt.Sprintf("\n    return string(m.Table().TableName)"))
	col = append(col, fmt.Sprintf("\n}\n"))

	switch req.DBType {
	case common.ConnectTypeClickhouse:
		col = append(col, fmt.Sprintf("\nfunc (m *%s) Table() *modao.ClickhouseTbl {", modName(req.TableName)))
		col = append(col, fmt.Sprintf("\n    return &modao.ClickhouseTbl{\n"))
	case common.ConnectTypeHologres:
		col = append(col, fmt.Sprintf("\nfunc (m *%s) Table() *modao.HologresTbl {", modName(req.TableName)))
		col = append(col, fmt.Sprintf("\n    return &modao.HologresTbl{\n"))
		col = append(col, fmt.Sprintf("\t\tPatternName: \"%s\",\n", req.PatternName))
	case common.ConnectTypeMysql:
		col = append(col, fmt.Sprintf("\nfunc (m *%s) Table() *modao.MysqlTbl {", modName(req.TableName)))
		col = append(col, fmt.Sprintf("\n    return &modao.MysqlTbl{\n"))
		col = append(col, fmt.Sprintf("\t\tConnectInformation: modao.ConnectInfo{\n"))
	default:
		panic("该数据库类型暂不支持,请重新选择数据库类型")
	}

	col = append(col, fmt.Sprintf("\t\t\tConnectName: \"%s\",\n\t\t},\n", req.DBName))
	col = append(col, fmt.Sprintf("\t\tDatabaseName: \"%s\",\n", req.DBName))
	col = append(col, fmt.Sprintf("\t\tTableName: \"%s\",\n", req.TableName))
	col = append(col, fmt.Sprintf("\t}\n"))
	col = append(col, fmt.Sprintf("}"))

	return
}

// 创建 dao struct 的方法
func createDaoMethods(req GenModDaoReq) (col []string) {
	table := req.TableName

	col = append(col, fmt.Sprintf("var %sSingle modao.SingleDao[*%s] \n\n", common.FirstToLower(daoName(table)), daoName(table)))
	col = append(col, fmt.Sprintf("func Instance%s(ctx context.Context) *%s {\n \treturn %sSingle.Do(ctx, func(withDebug bool) *%s {\n \t\treturn &%s{", daoName(table), daoName(table), common.FirstToLower(daoName(table)), daoName(table), daoName(table)))

	switch req.DBType {
	case common.ConnectTypeMysql:
		col = append(col, fmt.Sprintf("\n \t\t\tMysqlBaseDao: modao.NewMysqlBaseDao(&%s{}, withDebug),\n\t\t } \n\t }) \n} \n\n", modName(table)))
		col = append(col, fmt.Sprintf("type %s struct {\n \t*modao.MysqlBaseDao \n } \n\n", daoName(table)))
	case common.ConnectTypeHologres:
		col = append(col, fmt.Sprintf("\n \t\t\tBaseHgDao: INTComp.NewBaseHgDao(INTComp.NewHgOpMod(&%s{})),\n\t\t } \n\t }) \n} \n\n", modName(table)))
		col = append(col, fmt.Sprintf("type %s struct {\n \t*INTComp.BaseHgDao \n } \n\n", daoName(table)))
	default:
		panic("该数据库类型暂不支持,请重新选择数据库类型")
	}

	return
}

// GenAppendToFile 追加内容到文件
func GenAppendToFile(filename string, content string) (err error) {
	file, err := os.OpenFile(filename, os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return fmt.Errorf("无法打开已存在文件: %w", err)
	}
	defer file.Close()

	_, err = file.WriteString(content)
	if err != nil {
		return fmt.Errorf("写入文件失败: %w", err)
	}

	return
}
