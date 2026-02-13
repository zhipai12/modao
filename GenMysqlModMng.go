package modao

import (
	"fmt"
	"strings"

	ms "github.com/blastrain/vitess-sqlparser/sqlparser"
	"github.com/rrzu/cst"
)

type (
	// GenMysqlModMng 生成类 mysql mod 对象
	GenMysqlModMng struct {
		Req         GenModDaoReq // 请求参数
		Table       ModTable     // 表对象
		ConnectData ConnectData  // 连接信息
	}
)

func (g *GenMysqlModMng) Options() (rsp cst.Options[string, any], err error) {
	// 获取分析语法树对象
	var databases = g.databases()

	rsp = NamesToOptions[string, any](databases, cst.DataTypeString)
	if len(databases) == 0 {
		return
	}
	for i := range rsp.Opts {
		// 安全类型断言：确保Val是字符串类型
		if val, ok := rsp.Opts[i].Val.(string); ok {
			// 生成子选项（每次迭代创建新变量，取地址安全）
			optsVal := NamesToOptions[string, any](g.tables(val), cst.DataTypeString)
			// 将子选项指针赋值给 Sub 字段
			rsp.Opts[i].Sub = &optsVal
		}
	}

	return
}

func (g *GenMysqlModMng) SetGenModReq(req GenModDaoReq) {
	connectInfo := ConnectInfo{ConnectType: req.DBType}

	g.Req = req
	g.ConnectData.ConnectType = req.DBType
	g.ConnectData.Db = GetGormDb(connectInfo, true)
}

func (g *GenMysqlModMng) CreateTbl() (dest TblFromDb) {
	sql := fmt.Sprintf("show create table %s.%s", g.Req.DBName, g.Req.TableName)
	g.ConnectData.Db.Raw(sql).Take(&dest)
	return dest
}

func (g *GenMysqlModMng) GenerateModTable(createTbl TblFromDb) (tbl ModTable, err error) {
	defer func() { tbl = g.Table }()

	// 结构体字段
	stmt, er := ms.Parse(strings.TrimSpace(createTbl.CreateTable))
	if er != nil {
		return tbl, er
	}

	createTableObj, ok := stmt.(*ms.CreateTable)
	if !ok {
		return
	}

	g.Table.tableName = g.Req.TableName

	// 表注释
	for _, option := range createTableObj.Options {
		if option.Type == ms.TableOptionComment {
			g.Table.tableComment = option.StrValue
			break
		}
	}

	// 字段注释
	for _, column := range createTableObj.Columns {
		typ := g.convertTypeMs(column.Type)
		comment := ""
		for _, option := range column.Options {
			switch option.Type {
			case ms.ColumnOptionComment:
				comment = strings.Trim(option.Value, "\"")
			default:
			}
		}

		g.Table.SetColumnTyp(column.Name, typ).
			SetColumnTag(column.Name, toTag(column.Name)).
			SetColumnComment(column.Name, comment)
	}

	return
}

// ---------------------------
// ---       内部方法       ---
// ---------------------------

// 获取数据库
func (g *GenMysqlModMng) databases() (dest []string) {
	g.ConnectData.Db.Raw("show databases").Scan(&dest)
	return
}

// 获取表
func (g *GenMysqlModMng) tables(dbName string) (tables []string) {
	g.ConnectData.Db.Raw(fmt.Sprintf("show tables from %s", dbName)).Scan(&tables)
	return
}

// 处理字段类型显示
func (g *GenMysqlModMng) convertTypeMs(t string) string {
	t = strings.TrimSpace(strings.ToLower(t))
	var newType string

	if strings.Contains(t, "unsigned") {
		newType = "u" + strings.TrimSuffix(t, "unsigned")
	}

	switch {
	case strings.HasPrefix(t, "bigint"):
		newType = "int64"
	case strings.HasPrefix(t, "int") || strings.HasPrefix(t, "mediumint"):
		newType = "int"
	case strings.HasPrefix(t, "smallint"):
		newType = "int16"
	case strings.HasPrefix(t, "tinyint"):
		newType = "int8"
	case strings.HasPrefix(t, "decimal") || strings.HasPrefix(t, "float") || strings.HasPrefix(t, "numeric"):
		newType = "float64"
	case strings.HasPrefix(t, "datetime") || strings.HasPrefix(t, "timestamp"):
		newType = "time.Time"
	case strings.HasPrefix(t, "date"):
		newType = "int"
	case strings.HasPrefix(t, "char") || strings.HasPrefix(t, "varchar") || strings.HasPrefix(t, "text") || strings.HasPrefix(t, "mediumtext") || strings.HasPrefix(t, "longtext"):
		newType = "string"
	case strings.HasPrefix(t, "json"):
		newType = "string"
	case strings.HasPrefix(t, "binary") || strings.HasPrefix(t, "varbinary") || strings.HasPrefix(t, "blob") || strings.HasPrefix(t, "longblob"):
		newType = "[]byte"
	case strings.HasPrefix(t, "bool"):
		newType = "bool"
	case strings.HasPrefix(t, "time"):
		newType = "string"
	default:
		newType = "// todo"
	}

	return newType
}
