package common

const (
	ConnectTypeMysql      ConnectType = "mysql"      // mysql
	ConnectTypeHologres   ConnectType = "hologres"   // hologres
	ConnectTypeClickhouse ConnectType = "clickhouse" // clickhouse
	ConnectTypeMaxcompute ConnectType = "maxcompute" // maxcompute
)

type (
	ConnectType string // 连接类型
	ConnectName string // 连接名称
)

// ConnectTypeName 根据连接类型值返回其常量标识符名称
func ConnectTypeName(s ConnectType) string {
	switch s {
	case ConnectTypeMysql:
		return "ConnectTypeMysql"
	case ConnectTypeHologres:
		return "ConnectTypeHologres"
	case ConnectTypeClickhouse:
		return "ConnectTypeClickhouse"
	case ConnectTypeMaxcompute:
		return "ConnectTypeMaxcompute"
	default:
		return ""
	}
}
