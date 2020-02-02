package main

import (
	"database/sql"
	"github.com/jinzhu/gorm"
	"github.com/petrjahoda/zapsi_database"
	"time"
)

func UpdateDowntimeData(device zapsi_database.Device, id uint) {
	// TODO
}

func UpdateOrderData(device zapsi_database.Device, id uint) {
	// TODO
}

func OpenDowntime(device zapsi_database.Device, actualWorkplaceState zapsi_database.WorkplaceState, openOrderId uint) {
	connectionString, dialect := zapsi_database.CheckDatabaseType(DatabaseType, DatabaseIpAddress, DatabasePort, DatabaseLogin, DatabaseName, DatabasePassword)
	db, err := gorm.Open(dialect, connectionString)
	if err != nil {
		LogError(device.Name, "Problem opening "+DatabaseName+" database: "+err.Error())
		activeDevices = nil
		return
	}
	defer db.Close()
	var noReasonDowntime zapsi_database.Downtime
	db.Where("name = ?", "No reason downtime").Find(&noReasonDowntime)
	var downtimeToSave zapsi_database.DeviceDowntimeRecord
	downtimeToSave.DateTimeStart = actualWorkplaceState.DateTimeStart
	if openOrderId > 0 {
		downtimeToSave.DeviceOrderRecordId = sql.NullInt32{Int32: int32(openOrderId)}
		var deviceUserRecord zapsi_database.DeviceUserRecord
		db.Where("device_order_record_id = ?", openOrderId).Find(&deviceUserRecord)
		downtimeToSave.UserId = sql.NullInt32{Int32: int32(deviceUserRecord.UserId)}
	}
	downtimeToSave.Duration = time.Now().Sub(actualWorkplaceState.DateTimeStart)
	downtimeToSave.DeviceId = device.ID
	downtimeToSave.DowntimeId = noReasonDowntime.ID
	db.Save(&downtimeToSave)
}

func OpenOrder(device zapsi_database.Device, actualWorkplaceState zapsi_database.WorkplaceState) {
	connectionString, dialect := zapsi_database.CheckDatabaseType(DatabaseType, DatabaseIpAddress, DatabasePort, DatabaseLogin, DatabaseName, DatabasePassword)
	db, err := gorm.Open(dialect, connectionString)
	if err != nil {
		LogError(device.Name, "Problem opening "+DatabaseName+" database: "+err.Error())
		activeDevices = nil
		return
	}
	defer db.Close()
	var orderToSave zapsi_database.DeviceOrderRecord
	orderToSave.DateTimeStart = actualWorkplaceState.DateTimeStart
	orderToSave.Duration = time.Now().Sub(actualWorkplaceState.DateTimeStart)
	orderToSave.DeviceId = device.ID
	orderToSave.WorkplaceId = device.WorkplaceId
	db.Save(&orderToSave)
}

func CloseDowntime(device zapsi_database.Device, openDowntimeId uint) {
	connectionString, dialect := zapsi_database.CheckDatabaseType(DatabaseType, DatabaseIpAddress, DatabasePort, DatabaseLogin, DatabaseName, DatabasePassword)
	db, err := gorm.Open(dialect, connectionString)
	if err != nil {
		LogError(device.Name, "Problem opening "+DatabaseName+" database: "+err.Error())
		activeDevices = nil
		return
	}
	defer db.Close()
	var openDowntime zapsi_database.DeviceDowntimeRecord
	db.Where("id=?", openDowntimeId).Find(&openDowntime)
	openDowntime.DateTimeEnd = sql.NullTime{Time: time.Now()}
	openDowntime.Duration = time.Now().Sub(openDowntime.DateTimeStart)
	db.Save(&openDowntime)
}

func CloseOrder(device zapsi_database.Device, openOrderId uint) {
	connectionString, dialect := zapsi_database.CheckDatabaseType(DatabaseType, DatabaseIpAddress, DatabasePort, DatabaseLogin, DatabaseName, DatabasePassword)
	db, err := gorm.Open(dialect, connectionString)
	if err != nil {
		LogError(device.Name, "Problem opening "+DatabaseName+" database: "+err.Error())
		activeDevices = nil
		return
	}
	defer db.Close()
	var openOrder zapsi_database.DeviceOrderRecord
	db.Where("id=?", openOrderId).Find(&openOrder)
	openOrder.DateTimeEnd = sql.NullTime{Time: time.Now()}
	openOrder.Duration = time.Now().Sub(openOrder.DateTimeStart)
	db.Save(&openOrder)
}

func CheckOpenDowntime(device zapsi_database.Device) uint {
	connectionString, dialect := zapsi_database.CheckDatabaseType(DatabaseType, DatabaseIpAddress, DatabasePort, DatabaseLogin, DatabaseName, DatabasePassword)
	db, err := gorm.Open(dialect, connectionString)
	if err != nil {
		LogError(device.Name, "Problem opening "+DatabaseName+" database: "+err.Error())
		activeDevices = nil
		return 0
	}
	defer db.Close()
	var openDowntime zapsi_database.DeviceDowntimeRecord
	db.Where("device_id=?", device.ID).Where("date_time_end is null").Last(&openDowntime)
	return openDowntime.ID
}

func CheckOpenOrder(device zapsi_database.Device) uint {
	connectionString, dialect := zapsi_database.CheckDatabaseType(DatabaseType, DatabaseIpAddress, DatabasePort, DatabaseLogin, DatabaseName, DatabasePassword)
	db, err := gorm.Open(dialect, connectionString)
	if err != nil {
		LogError(device.Name, "Problem opening "+DatabaseName+" database: "+err.Error())
		activeDevices = nil
		return 0
	}
	defer db.Close()
	var openOrder zapsi_database.DeviceOrderRecord
	db.Where("device_id=?", device.ID).Where("date_time_end is null").Last(&openOrder)
	return openOrder.ID
}

func GetActualState(device zapsi_database.Device) (zapsi_database.State, zapsi_database.WorkplaceState) {
	connectionString, dialect := zapsi_database.CheckDatabaseType(DatabaseType, DatabaseIpAddress, DatabasePort, DatabaseLogin, DatabaseName, DatabasePassword)
	db, err := gorm.Open(dialect, connectionString)
	if err != nil {
		LogError(device.Name, "Problem opening "+DatabaseName+" database: "+err.Error())
		activeDevices = nil
		return zapsi_database.State{}, zapsi_database.WorkplaceState{}
	}
	defer db.Close()
	var workplaceState zapsi_database.WorkplaceState
	db.Where("workplace_id=?", device.WorkplaceId).Last(&workplaceState)
	var actualState zapsi_database.State
	db.Where("id=?", workplaceState.StateId).Last(&actualState)
	return actualState, workplaceState

}
