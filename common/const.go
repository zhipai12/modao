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
