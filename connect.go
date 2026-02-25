package modao

import (
	"sync"

	"github.com/rrzu/modao/common"
	"gorm.io/gorm"
)

var gormConnectContainer sync.Map

// ConnectInfo 连接信息
type ConnectInfo struct {
	ConnectName common.ConnectName
	ConnectType common.ConnectType
}

// ConnectData 连接数据（保持原名）
type ConnectData struct {
	Db          *gorm.DB
	ConnectInfo ConnectInfo
}

// RegisterGormDb 注册连接（已存在则忽略）
func RegisterGormDb(info ConnectInfo, db *gorm.DB) {
	if info.ConnectName == "" {
		return
	}
	if _, exists := gormConnectContainer.Load(key(info)); !exists {
		gormConnectContainer.Store(key(info), &ConnectData{
			Db: db,
			ConnectInfo: ConnectInfo{
				ConnectName: info.ConnectName,
				ConnectType: info.ConnectType,
			},
		})
	}
}

// ModifyGormDb 强制覆盖更新连接
func ModifyGormDb(info ConnectInfo, db *gorm.DB) {
	if info.ConnectName == "" {
		return
	}

	gormConnectContainer.Store(key(info), &ConnectData{
		Db: db,
		ConnectInfo: ConnectInfo{
			ConnectName: info.ConnectName,
			ConnectType: info.ConnectType,
		},
	})
}

// GetGormDb 获取连接
func GetGormDb(info ConnectInfo, withDebug bool) (gormDB *gorm.DB) {
	defer func() {
		if withDebug && gormDB != nil {
			gormDB = gormDB.Debug()
		}
	}()

	if info.ConnectName != "" {
		if v, ok := gormConnectContainer.Load(key(info)); ok {
			if data, ok := v.(*ConnectData); ok {
				return data.Db
			}
		}
		return nil
	}

	// ConnectName 为空 → 按类型取首个
	var found *ConnectData
	gormConnectContainer.Range(func(_, val interface{}) bool {
		if data, ok := val.(*ConnectData); ok && data.ConnectInfo.ConnectType == info.ConnectType {
			found = data
			return false
		}
		return true
	})

	if found != nil {
		return found.Db
	}

	return nil
}

// GetAllConnectData 获取所有连接数据
func GetAllConnectData() (res []ConnectData) {
	gormConnectContainer.Range(func(_, val interface{}) bool {
		if data, ok := val.(*ConnectData); ok {
			res = append(res, *data)
		}
		return true
	})

	return
}

func key(info ConnectInfo) string {
	return string(info.ConnectName)
}
