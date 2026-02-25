package main

import (
	"github.com/rrzu/modao"
	"github.com/rrzu/modao/common"
	"gorm.io/gorm"
)

func main() {
	(&modao.Config{}).
		SetDebugKey("onDebug").
		SetGormDb(modao.ConnectInfo{ConnectName: "mysql-sz-dev", ConnectType: common.ConnectTypeMysql}, &gorm.DB{}).
		SetGormDb(modao.ConnectInfo{ConnectName: "hologres-sz-dev", ConnectType: common.ConnectTypeHologres}, &gorm.DB{}).
		SetGenMdPath("/mnt/d/work/go/go_package/modao").
		Init()

	// ctx := context.Background()
	//
	// dao := repository.InstanceAccountServerMapDao(ctx)
	// mod := dao.Mod()
	// mod.Table()
	//
	// dao2 := repository.InstanceAccountDao(ctx)
	// dao2.Db()
	// mod2 := dao2.Mod()
	// mod2.Table()
	//
	// dao3 := repository.InstanceDimOrderTypeDao(ctx)
	// dao3.Qry()
	// mod3 := dao3.Mod()
	// mod3.Table()
	//
	// sql := dao3.Db().ToSQL(func(tx *gorm.DB) *gorm.DB {
	// 	return tx.Select("*").Find(&map[string]interface{}{})
	// })
	//
	// 输出：SELECT * FROM `dim_order_type`
	// fmt.Println(sql)
}
