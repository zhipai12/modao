package modao

import (
	"errors"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
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
	switch connect.ConnectInfo.ConnectType {
	case common.ConnectTypeMysql:
		mng = &GenMysqlModMng{
			ConnectData: connect,
		}
	// case _const.DBTypeCh:
	// case _const.DBTypePg:
	case common.ConnectTypeHologres:
		mng = &GenHologresModMng{
			ConnectData: connect,
		}
	default:
		err = fmt.Errorf("错误类型")
	}
	return
}

// GenConstContent 生成常量内容
func GenConstContent(req GenModDaoReq) (err error) {
	var (
		partPath    string
		constName   string
		constVal    string
		constType   string
		packageName string
	)

	switch req.DBType {
	case common.ConnectTypeMysql:
		partPath = filepath.Join("ms")
		constName = common.SnakeToPascal(req.DBName)
		constVal = req.DBName
		constType = "modao.MysqlDatabaseName"
		packageName = "ms"
	case common.ConnectTypeHologres:
		partPath = filepath.Join("hg", "data_analysis")
		constName = common.SnakeToPascal(req.PatternName)
		constVal = req.PatternName
		constType = "modao.HologresPatternName"
		packageName = "data_analysis"
	default:
		return errors.New("该数据库类型暂不支持，请重新选择数据库类型")
	}

	filePath := filepath.Join(OutputPath, partPath, "const_gen.go")
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		return fmt.Errorf("获取绝对路径失败: %w", err)
	}

	exists := common.CheckFileIsExist(absPath)
	if exists {
		// 文件已存在, 检查添加常量
		err = EnsureConstInFile(absPath, constName, constType, constVal)
		if err != nil {
			return fmt.Errorf("确保常量失败: %w", err)
		}
		return nil
	}

	// 文件不存在：创建并写入初始内容
	if err = common.CreateFile(absPath); err != nil {
		return fmt.Errorf("创建文件失败: %w", err)
	}

	// 4.2 打开文件准备写入
	file, err := os.OpenFile(absPath, os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		return fmt.Errorf("打开文件失败: %w", err)
	}
	defer file.Close()

	content := fmt.Sprintf("package %s\n\nimport \"github.com/rrzu/modao\"\n\nconst %s %s = \"%s\"\n",
		packageName, constName, constType, constVal)

	if _, err = file.WriteString(content); err != nil {
		return fmt.Errorf("写入文件内容失败: %w", err)
	}

	return nil
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
	filename = filepath.Join(OutputPath, partPath, common.UnderScoreToCamel(req.TableName)+"_gen.go")

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
	connect.ConnectInfo.ConnectType = req.DBType
	connect.ConnectInfo.ConnectName = req.ConnectName

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
	tableName := req.TableName
	modelName := modName(tableName)
	pascalTable := common.SnakeToPascal(tableName)

	// TableName 方法
	col = append(col, fmt.Sprintf("\nfunc (m *%s) TableName() string {", modelName))
	col = append(col, fmt.Sprintf("\n    return string(m.Table().TableName)"))
	col = append(col, fmt.Sprintf("\n}\n"))

	// Table 方法头（根据数据库类型不同）
	switch req.DBType {
	case common.ConnectTypeHologres:
		col = append(col, fmt.Sprintf("\nfunc (m *%s) Table() *modao.HologresTbl {", modelName))
		col = append(col, fmt.Sprintf("\n    return &modao.HologresTbl{\n"))
		col = append(col, fmt.Sprintf("\t\tPatternName: data_analysis.%s,\n", common.SnakeToPascal(req.PatternName))) // Hologres 独有字段
	case common.ConnectTypeMysql:
		col = append(col, fmt.Sprintf("\nfunc (m *%s) Table() *modao.MysqlTbl {", modelName))
		col = append(col, fmt.Sprintf("\n    return &modao.MysqlTbl{\n"))
		col = append(col, fmt.Sprintf("\t\tDatabaseName: ms.%s,\n", common.SnakeToPascal(req.DBName)))
	default:
		panic("该数据库类型暂不支持,请重新选择数据库类型")
	}

	// 公共字段部分
	col = append(col, fmt.Sprintf("\t\tConnectInformation: modao.ConnectInfo{\n"))
	col = append(col, fmt.Sprintf("\t\t\tConnectType: common.%s,\n", common.ConnectTypeName(req.DBType)))
	col = append(col, fmt.Sprintf("\t\t\tConnectName: \"%s\",\n", req.ConnectName))
	col = append(col, fmt.Sprintf("\t\t},\n"))
	col = append(col, fmt.Sprintf("\t\tTableName: %s,\n", pascalTable))
	col = append(col, fmt.Sprintf("\t}\n"))
	col = append(col, fmt.Sprintf("}"))

	return
}

// 创建 dao struct 的方法
// createDaoMethods 生成 dao struct 的方法
func createDaoMethods(req GenModDaoReq) (col []string) {
	tableName := req.TableName
	pascalTable := common.SnakeToPascal(tableName)
	daoNameStr := daoName(tableName)
	modelNameStr := modName(tableName)

	// 常量定义（根据数据库类型不同）
	var constLine string
	switch req.DBType {
	case common.ConnectTypeMysql:
		constLine = fmt.Sprintf("const %s modao.MysqlTableName = \"%s\"\n\n", pascalTable, tableName)
	case common.ConnectTypeHologres:
		constLine = fmt.Sprintf("const %s modao.HologresTableName = \"%s\"\n\n", pascalTable, tableName)
	default:
		panic("该数据库类型暂不支持,请重新选择数据库类型")
	}

	col = append(col, constLine)

	// SingleDao 变量声明
	col = append(col, fmt.Sprintf("var %sSingle modao.SingleDao[*%s] \n\n", common.FirstToLower(daoNameStr), daoNameStr))

	// Instance 函数开始部分
	instanceStart := fmt.Sprintf("func Instance%s(ctx context.Context) *%s {\n \treturn %sSingle.Do(ctx, func(withDebug bool) *%s {\n \t\treturn &%s{",
		daoNameStr, daoNameStr, common.FirstToLower(daoNameStr), daoNameStr, daoNameStr)
	col = append(col, instanceStart)

	// 根据数据库类型生成基础 Dao 初始化和结构体定义
	switch req.DBType {
	case common.ConnectTypeMysql:
		col = append(col, fmt.Sprintf("\n \t\t\tMysqlBaseDao: modao.NewMysqlBaseDao(&%s{}, withDebug),\n\t\t } \n\t }) \n} \n\n", modelNameStr))
		col = append(col, fmt.Sprintf("type %s struct {\n \t*modao.MysqlBaseDao \n } \n\n", daoNameStr))
	case common.ConnectTypeHologres:
		col = append(col, fmt.Sprintf("\n \t\t\tHologresBaseDao: modao.NewHologresBaseDao(&%s{}, withDebug),\n\t\t } \n\t }) \n} \n\n", modelNameStr))
		col = append(col, fmt.Sprintf("type %s struct {\n \t*modao.HologresBaseDao \n } \n\n", daoNameStr))
	default:
		panic("该数据库类型暂不支持,请重新选择数据库类型")
	}

	return
}

// GenAppendToFile 追加内容到文件
func GenAppendToFile(filename string, content string) (err error) {
	file, err := os.OpenFile(filename, os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		fmt.Printf("错误 %s", err)
		return fmt.Errorf("无法打开已存在文件: %w", err)
	}
	defer file.Close()

	_, err = file.WriteString(content)
	if err != nil {
		fmt.Printf("错误 %s", err)
		return fmt.Errorf("写入文件失败: %w", err)
	}

	return
}

// EnsureConstInFile 检查指定 Go 文件中是否存在常量 constName。
func EnsureConstInFile(filename, constName, constType, constValue string) error {
	fileSet := token.NewFileSet()
	astFile, err := parser.ParseFile(fileSet, filename, nil, parser.ParseComments)
	if err != nil {
		return err
	}

	// 遍历 AST 查找常量
	for _, decl := range astFile.Decls {
		genDecl, ok := decl.(*ast.GenDecl)
		if !ok || genDecl.Tok != token.CONST {
			continue
		}
		for _, spec := range genDecl.Specs {
			for _, name := range spec.(*ast.ValueSpec).Names {
				if name.Name == constName {
					return nil // 已存在
				}
			}
		}
	}

	// 追加常量定义（前置换行避免粘连）
	file, err := os.OpenFile(filename, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()
	_, err = file.WriteString(fmt.Sprintf("\nconst %s %s = \"%s\"", constName, constType, constValue))
	return err
}
