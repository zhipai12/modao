package modao

import (
	"fmt"
	"log"
	"strings"

	pg "github.com/pganalyze/pg_query_go/v2"
	"github.com/rrzu/cst"
)

type (
	// GenHologresModMng 生成类 Hologres mod 对象
	GenHologresModMng struct {
		Req         GenModDaoReq // 请求参数
		Table       ModTable     // 表对象
		ConnectData ConnectData  // 连接信息
	}
)

func (g *GenHologresModMng) Options() (rsp cst.Options[string, any], err error) {
	// 获取分析语法树对象
	databases := g.databases()

	rsp = NamesToOptions[string, any](databases, cst.DataTypeString)
	if len(databases) == 0 {
		return
	}

	for i := range rsp.Opts {
		subOptI := NamesToOptions[string, any](g.getPattern(), cst.DataTypeString)
		rsp.Opts[i].Sub = &subOptI

		for i2, opts := range rsp.Opts[i].Sub.Opts {
			subOptI2 := NamesToOptions[string, any](g.tables(opts.Val.(string)), cst.DataTypeString)
			rsp.Opts[i].Sub.Opts[i2].Sub = &subOptI2
		}
	}

	return
}

func (g *GenHologresModMng) SetGenModReq(req GenModDaoReq) {
	connectInfo := ConnectInfo{
		ConnectType: req.DBType,
		ConnectName: req.ConnectName,
	}

	g.Req = req
	g.ConnectData.ConnectInfo = connectInfo
	g.ConnectData.Db = GetGormDb(connectInfo, true)
}

func (g *GenHologresModMng) CreateTbl() (dest TblFromDb) {
	sql := fmt.Sprintf(`
		WITH columns AS (
			SELECT
				a.attnum,
				'    ' || a.attname || ' ' ||
				pg_catalog.format_type(a.atttypid, a.atttypmod) ||
				CASE WHEN a.attnotnull THEN ' NOT NULL' ELSE '' END ||
				COALESCE(' DEFAULT ' || (
					SELECT pg_get_expr(d.adbin, d.adrelid)
					FROM pg_attrdef d
					WHERE d.adrelid = a.attrelid AND d.adnum = a.attnum
				), '') AS col_def
			FROM pg_attribute a
			WHERE a.attrelid = '%s.%s'::regclass
			  AND a.attnum > 0
			  AND NOT a.attisdropped
			ORDER BY a.attnum
		)
		SELECT
			'CREATE TABLE ' || n.nspname || '.' || c.relname || ' (' || chr(10) ||
			string_agg(col_def, ',' || chr(10) ORDER BY attnum) ||
			chr(10) || ');' || chr(10) || chr(10) ||

			-- 表注释（可为空）
			COALESCE(
				'COMMENT ON TABLE ' || n.nspname || '.' || c.relname ||
				' IS ''' || replace(obj_description(c.oid), '''', '''''') || ''';' || chr(10),
				''
			) ||

			-- 字段注释（可为空）
			COALESCE(
				(SELECT string_agg(
					'COMMENT ON COLUMN ' || n.nspname || '.' || c.relname || '.' || a.attname ||
					' IS ''' || replace(col_description(a.attrelid, a.attnum), '''', '''''') || ''';' || chr(10),
					''
				)
				FROM pg_attribute a
				WHERE a.attrelid = c.oid
				  AND a.attnum > 0
				  AND col_description(a.attrelid, a.attnum) IS NOT NULL
				),
				''
			) AS "Create Table"

		FROM pg_class c
		JOIN pg_namespace n ON n.oid = c.relnamespace
		CROSS JOIN columns
		WHERE c.relname = '%s'
		  AND n.nspname = '%s'
		GROUP BY n.nspname, c.relname, c.oid
	`,
		g.Req.PatternName, g.Req.TableName,
		g.Req.TableName, g.Req.PatternName)

	g.ConnectData.Db.Raw(sql).Take(&dest)

	return dest
}

func (g *GenHologresModMng) GenerateModTable(createTbl TblFromDb) (tbl ModTable, err error) {
	defer func() { tbl = g.Table }()

	result, er := pg.Parse(createTbl.CreateTable)
	if er != nil {
		return tbl, er
	}

	// 第一次遍历：收集所有注释信息
	for _, stmt := range result.Stmts {
		// 表名
		if createStmt := stmt.Stmt.GetCreateStmt(); createStmt != nil {
			g.Table.tableName = createStmt.Relation.Relname
		}

		// 表注释、列注释
		if commentStmt := stmt.Stmt.GetCommentStmt(); commentStmt != nil {
			// 表注释
			if commentStmt.Objtype == pg.ObjectType_OBJECT_TABLE {
				g.Table.tableComment = commentStmt.Comment
			}

			// 列注释
			if commentStmt.Objtype == pg.ObjectType_OBJECT_COLUMN {
				// 解析格式：public.employees.name → employees.name
				items := commentStmt.Object.GetList().GetItems()
				if items == nil || len(items) < 3 {
					continue
				}

				field := items[2].GetString_().Str
				g.Table.SetColumnComment(field, commentStmt.Comment)
			}
		}
	}

	// 第二次遍历：处理表结构
	for _, stmt := range result.Stmts {
		createStmt := stmt.Stmt.GetCreateStmt()
		if createStmt == nil {
			continue
		}

		for _, elt := range createStmt.TableElts {
			column := elt.GetColumnDef()
			if column == nil {
				continue
			}

			// 打印字段信息
			g.Table.
				SetColumnTyp(column.Colname, g.getRawFieldType(column.TypeName)).
				SetColumnTag(column.Colname, toTag(column.Colname))
		}
	}
	return
}

// ---------------------------
// ---       内部方法       ---
// ---------------------------

// 获取数据库
func (g *GenHologresModMng) databases() (dbNames []string) {
	dbNames = make([]string, 0)
	dbNames = append(dbNames, "data_analysis_test")

	g.ConnectData.Db.Raw(`SELECT datname FROM pg_database`).Pluck("datname", &dbNames)

	return
}

// 获取模式
func (g *GenHologresModMng) getPattern() (patterns []string) {
	patterns = make([]string, 0)

	sql := `SELECT nspname AS schema_name FROM pg_namespace WHERE nspowner > 0 AND nspname NOT IN ('information_schema', 'hg_catalog')`
	g.ConnectData.Db.Raw(sql).Pluck("nspname", &patterns)

	return
}

// 获取表
func (g *GenHologresModMng) tables(patternName string) (tables []string) {
	tables = make([]string, 0)
	g.ConnectData.Db.Raw(fmt.Sprintf(`SELECT tablename FROM pg_tables WHERE schemaname = '%s'`, patternName)).Pluck("tablename", &tables)
	return
}

// 字段类型
func (g *GenHologresModMng) getRawFieldType(typeName *pg.TypeName) string {
	var typeParts []string
	for _, name := range typeName.Names {
		part := name.GetString_().Str
		if part != "pg_catalog" && part != "public" { // 过滤常见 schema
			typeParts = append(typeParts, part)
		}
	}
	dbType := strings.ToLower(strings.Join(typeParts, "."))

	switch dbType {
	case "int8", "bigint":
		return "int64"
	case "int4", "integer":
		return "int32"
	case "int2", "smallint":
		return "int16"
	case "text", "varchar", "char", "bpchar", "name", "character varying":
		return "string"
	case "numeric", "decimal":
		return "string"
	case "float8", "double precision":
		return "float64"
	case "float4", "real":
		return "float32"
	case "bool", "boolean":
		return "bool"
	case "timestamp", "timestamp without time zone":
		return "time.Time"
	case "timestamptz", "timestamp with time zone":
		return "time.Time"
	case "date":
		return "time.Time"
	case "roaringbitmap":
		return "string" // 自定义类型，默认转为 string
	default:
		// 日志提醒未知类型
		log.Printf("unknown db type: %s", dbType)
		return "string"
	}
}
