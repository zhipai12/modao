package modao

import (
	"reflect"

	"github.com/pkg/errors"
	"github.com/rrzu/modao/common"
	"gorm.io/gorm"
)

// IGormDb 获取数据库查询对象
type IGormDb interface {
	Db(obj ...*gorm.DB) *gorm.DB
}

// ---------------------------
// ----      BaseDao      ----
// ---------------------------

type BaseDao[T ITbl] struct {
	db  *gorm.DB
	mod IMod[T]
}

func NewBaseDao[T ITbl](m IMod[T], withDebug bool) *BaseDao[T] {
	d := new(BaseDao[T])
	d.db = GetGormDb(m.Table().ConnectInfo(), withDebug)
	d.mod = m
	return d
}

func (d *BaseDao[T]) Db(obj ...*gorm.DB) *gorm.DB {
	return d.db
}

func (d *BaseDao[T]) Qry() *gorm.DB {
	return d.Db().Table(d.Mod().Table().QueryTableName())
}

func (d *BaseDao[T]) Mod() IMod[T] {
	return d.mod
}

func (d *BaseDao[T]) ModTabQueryTableName() string {
	return d.Mod().Table().QueryTableName()
}

// BatchInsert 批量插入数据
func (d *BaseDao[T]) BatchInsert(data interface{}, size int) error {
	if size == 0 {
		return nil
	}
	batches := d.Db().Table(d.ModTabQueryTableName()).CreateInBatches(data, size)
	return batches.Error
}

// AutoBatchInsert 自动分批再批量插入数据
func (d *BaseDao[T]) AutoBatchInsert(data interface{}, batch int) (err error) {
	if batch <= 0 {
		return errors.Errorf("批量需要大于0")
	}

	rv := reflect.ValueOf(data)
	if rv.Kind() != reflect.Slice && rv.Kind() != reflect.Array {
		return errors.Errorf("不是一个有效的切片或数组类型")
	}

	// 数据长度
	length := rv.Len()
	// 分批批数
	times := length/batch + common.TernaryAny(length%batch > 0, 1, 0)
	// 本次插入数据长度
	thisBatchLength := batch

	for n := 0; n < times; n++ {
		cursorLeft := batch * n
		cursorRight := (n + 1) * batch

		// 最后一批数据
		if cursorRight > length {
			cursorRight = length
			thisBatchLength = cursorRight - cursorLeft
		}

		thisBatch := rv.Slice(cursorLeft, cursorRight).Interface()
		thisTx := d.Qry().CreateInBatches(thisBatch, thisBatchLength)
		if thisTx.Error != nil {
			err = thisTx.Error
			return
		}
	}

	return
}
