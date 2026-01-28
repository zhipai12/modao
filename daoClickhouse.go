package modao

type ClickhouseBaseDao struct {
	*BaseDao[*ClickhouseTbl]
}

func NewClickhouseBaseDao(mod IMod[*ClickhouseTbl], withDebug bool) *ClickhouseBaseDao {
	return &ClickhouseBaseDao{
		NewBaseDao(mod, withDebug),
	}
}

func (ch *ClickhouseBaseDao) Mod() IMod[*ClickhouseTbl] {
	return ch.mod.(IMod[*ClickhouseTbl])
}

// ListTableColumns 获取表字段
func (ch *ClickhouseBaseDao) ListTableColumns() (res []map[string]interface{}) {
	handle := ch.Db().Table("system.columns").Select("name, type")

	if len(ch.Mod().Table().DatabaseName) > 0 {
		handle.Where("database = ?", ch.Mod().Table().DatabaseName)
	}

	handle.Where("table = ?", ch.Mod().TableName()).Find(&res)
	return
}
