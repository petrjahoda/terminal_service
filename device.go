package main

import (
	"database/sql"
	"github.com/jinzhu/gorm"
	"github.com/petrjahoda/zapsi_database"
	"time"
)

func UpdateDowntimeData(device zapsi_database.Device, deviceDowntimeRecordId uint) {
	connectionString, dialect := zapsi_database.CheckDatabaseType(DatabaseType, DatabaseIpAddress, DatabasePort, DatabaseLogin, DatabaseName, DatabasePassword)
	db, err := gorm.Open(dialect, connectionString)
	if err != nil {
		LogError(device.Name, "Problem opening "+DatabaseName+" database: "+err.Error())
		activeDevices = nil
		return
	}
	defer db.Close()
	var openDowntime zapsi_database.DownTimeRecord
	db.Where("id=?", deviceDowntimeRecordId).Find(&openDowntime)
	openDowntime.Duration = time.Now().Sub(openDowntime.DateTimeStart)
	db.Save(&openDowntime)
}

func UpdateOrderData(device zapsi_database.Device, deviceOrderRecordId uint) {
	connectionString, dialect := zapsi_database.CheckDatabaseType(DatabaseType, DatabaseIpAddress, DatabasePort, DatabaseLogin, DatabaseName, DatabasePassword)
	db, err := gorm.Open(dialect, connectionString)
	if err != nil {
		LogError(device.Name, "Problem opening "+DatabaseName+" database: "+err.Error())
		activeDevices = nil
		return
	}
	defer db.Close()
	var openOrder zapsi_database.OrderRecord
	db.Where("id=?", deviceOrderRecordId).Find(&openOrder)
	var workplacePortOk zapsi_database.WorkplacePort
	db.Where("workplace_id = ?", device.WorkplaceId).Where("counter_ok is true").Find(&workplacePortOk)
	countOk := 0
	db.Model(&zapsi_database.DevicePortDigitalRecord{}).Where("device_port_id = ?", workplacePortOk.DevicePortId).Where("date_time>?", openOrder.DateTimeStart).Where("data = 0").Count(&countOk)
	var workplacePortNok zapsi_database.WorkplacePort
	db.Where("workplace_id = ?", device.WorkplaceId).Where("counter_nok is true").Find(&workplacePortNok)
	countNok := 0
	db.Model(&zapsi_database.DevicePortDigitalRecord{}).Where("device_port_id = ?", workplacePortNok.DevicePortId).Where("date_time>?", openOrder.DateTimeStart).Where("data = 0").Count(&countNok)
	averageCycle := 0.0
	if countOk > 0 {
		averageCycle = time.Now().Sub(openOrder.DateTimeStart).Seconds() / float64(countOk)
	}
	openOrder.AverageCycle = float32(averageCycle)
	openOrder.CountOk = uint(countOk)
	openOrder.CountNok = uint(countNok)
	openOrder.Duration = time.Now().Sub(openOrder.DateTimeStart)
	openOrder.WorkplaceId = device.WorkplaceId
	db.Save(&openOrder)
}

func OpenDowntime(device zapsi_database.Device, actualWorkplaceState zapsi_database.StateRecord, openOrderId uint) {
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
	var downtimeToSave zapsi_database.DownTimeRecord
	downtimeToSave.DateTimeStart = actualWorkplaceState.DateTimeStart
	if openOrderId > 0 {
		downtimeToSave.OrderRecordId = sql.NullInt32{Int32: int32(openOrderId)}
		var deviceUserRecord zapsi_database.UserRecord
		db.Where("device_order_record_id = ?", openOrderId).Find(&deviceUserRecord)
		downtimeToSave.UserId = sql.NullInt32{Int32: int32(deviceUserRecord.UserId)}
	}
	downtimeToSave.Duration = time.Now().Sub(actualWorkplaceState.DateTimeStart)
	downtimeToSave.DeviceId = device.ID
	downtimeToSave.DowntimeId = noReasonDowntime.ID
	db.Save(&downtimeToSave)
}

func OpenOrder(device zapsi_database.Device, actualWorkplaceState zapsi_database.StateRecord) {
	connectionString, dialect := zapsi_database.CheckDatabaseType(DatabaseType, DatabaseIpAddress, DatabasePort, DatabaseLogin, DatabaseName, DatabasePassword)
	db, err := gorm.Open(dialect, connectionString)
	if err != nil {
		LogError(device.Name, "Problem opening "+DatabaseName+" database: "+err.Error())
		activeDevices = nil
		return
	}
	defer db.Close()
	var orderToSave zapsi_database.OrderRecord
	orderToSave.DateTimeStart = actualWorkplaceState.DateTimeStart
	orderToSave.Duration = time.Now().Sub(actualWorkplaceState.DateTimeStart)
	orderToSave.DeviceId = device.ID
	orderToSave.WorkplaceId = device.WorkplaceId
	orderToSave.Cavity = 1
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
	var openDowntime zapsi_database.DownTimeRecord
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
	var openOrder zapsi_database.OrderRecord
	db.Where("id=?", openOrderId).Find(&openOrder)
	var workplacePortOk zapsi_database.WorkplacePort
	db.Where("workplace_id = ?", device.WorkplaceId).Where("counter_ok is true").Find(&workplacePortOk)
	countOk := 0
	db.Model(&zapsi_database.DevicePortDigitalRecord{}).Where("device_port_id = ?", workplacePortOk.DevicePortId).Where("date_time>?", openOrder.DateTimeStart).Where("data = 0").Count(&countOk)
	averageCycle := 0.0
	if countOk > 0 {
		averageCycle = time.Now().Sub(openOrder.DateTimeStart).Seconds() / float64(countOk)
	}
	openOrder.AverageCycle = float32(averageCycle)
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
	var openDowntime zapsi_database.DownTimeRecord
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
	var openOrder zapsi_database.OrderRecord
	db.Where("device_id=?", device.ID).Where("date_time_end is null").Last(&openOrder)
	return openOrder.ID
}

func GetActualState(device zapsi_database.Device) (zapsi_database.State, zapsi_database.StateRecord) {
	connectionString, dialect := zapsi_database.CheckDatabaseType(DatabaseType, DatabaseIpAddress, DatabasePort, DatabaseLogin, DatabaseName, DatabasePassword)
	db, err := gorm.Open(dialect, connectionString)
	if err != nil {
		LogError(device.Name, "Problem opening "+DatabaseName+" database: "+err.Error())
		activeDevices = nil
		return zapsi_database.State{}, zapsi_database.StateRecord{}
	}
	defer db.Close()
	var workplaceState zapsi_database.StateRecord
	db.Where("workplace_id=?", device.WorkplaceId).Last(&workplaceState)
	var actualState zapsi_database.State
	db.Where("id=?", workplaceState.StateId).Last(&actualState)
	return actualState, workplaceState

}
