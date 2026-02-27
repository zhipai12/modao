package modao

type HologresBaseDao struct {
	*BaseDao[*HologresTbl]
}

func NewHologresBaseDao(mod IMod[*HologresTbl], withDebug bool) *HologresBaseDao {
	return &HologresBaseDao{
		NewBaseDao(mod, withDebug),
	}
}

func (ch *HologresBaseDao) Mod() IMod[*HologresTbl] {
	return ch.mod.(IMod[*HologresTbl])
}

func (ch *HologresBaseDao) GetDatabaseName() (name DatabaseName) {
	ch.Db().Raw("SELECT current_database()").Scan(&name)
	return
}
