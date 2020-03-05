package main

import (
	"database/sql"
	"github.com/jinzhu/gorm"
	"github.com/petrjahoda/zapsi_database"
	"time"
)

func UpdateDowntimeData(device zapsi_database.Device, deviceDowntimeRecordId int) {
	LogInfo(device.Name, "Updating downtime data")
	timer := time.Now()
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
	LogInfo(device.Name, "Downtime data updated, elapsed: "+time.Since(timer).String())
}

func UpdateOrderData(device zapsi_database.Device, deviceOrderRecordId int) {
	LogInfo(device.Name, "Updating order data")
	timer := time.Now()
	connectionString, dialect := zapsi_database.CheckDatabaseType(DatabaseType, DatabaseIpAddress, DatabasePort, DatabaseLogin, DatabaseName, DatabasePassword)
	db, err := gorm.Open(dialect, connectionString)
	if err != nil {
		LogError(device.Name, "Problem opening "+DatabaseName+" database: "+err.Error())
		activeDevices = nil
		return
	}
	defer db.Close()
	var deviceWorkplaceRecord zapsi_database.DeviceWorkplaceRecord
	db.Where("device_id = ?", device.ID).Find(&deviceWorkplaceRecord)
	var openOrder zapsi_database.OrderRecord
	db.Where("id=?", deviceOrderRecordId).Find(&openOrder)
	var workplacePortOk zapsi_database.WorkplacePort
	db.Where("workplace_id = ?", deviceWorkplaceRecord.WorkplaceID).Where("counter_ok = ?", "1").Find(&workplacePortOk)
	countOk := 0
	db.Model(&zapsi_database.DevicePortDigitalRecord{}).Where("device_port_id = ?", workplacePortOk.DevicePortId).Where("date_time>?", openOrder.DateTimeStart).Where("data = 0").Count(&countOk)
	var workplacePortNok zapsi_database.WorkplacePort
	db.Where("workplace_id = ?", deviceWorkplaceRecord.WorkplaceID).Where("counter_nok = ?", "1").Find(&workplacePortNok)
	countNok := 0
	db.Model(&zapsi_database.DevicePortDigitalRecord{}).Where("device_port_id = ?", workplacePortNok.DevicePortId).Where("date_time>?", openOrder.DateTimeStart).Where("data = 0").Count(&countNok)
	averageCycle := 0.0
	if countOk > 0 {
		averageCycle = time.Now().Sub(openOrder.DateTimeStart).Seconds() / float64(countOk)
	}
	openOrder.AverageCycle = float32(averageCycle)
	openOrder.CountOk = countOk
	openOrder.CountNok = countNok
	openOrder.Duration = time.Now().Sub(openOrder.DateTimeStart)
	openOrder.WorkplaceId = deviceWorkplaceRecord.WorkplaceID
	db.Save(&openOrder)
	LogInfo(device.Name, "Order data updated, elapsed: "+time.Since(timer).String())

}

func OpenDowntime(device zapsi_database.Device, actualWorkplaceState zapsi_database.StateRecord, openOrderId int) {
	LogInfo(device.Name, "Opening downtime")
	timer := time.Now()
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
	var deviceWorkplaceRecord zapsi_database.DeviceWorkplaceRecord
	db.Where("device_id = ?", device.ID).Find(&deviceWorkplaceRecord)
	var downtimeToSave zapsi_database.DownTimeRecord
	downtimeToSave.DateTimeStart = actualWorkplaceState.DateTimeStart
	if openOrderId > 0 {
		downtimeToSave.OrderRecordId = sql.NullInt32{Int32: int32(openOrderId)}
		var deviceUserRecord zapsi_database.UserRecord
		db.Where("order_record_id = ?", openOrderId).Find(&deviceUserRecord)
		downtimeToSave.UserId = sql.NullInt32{Int32: int32(deviceUserRecord.UserId)}
	}
	downtimeToSave.Duration = time.Now().Sub(actualWorkplaceState.DateTimeStart)
	downtimeToSave.DeviceId = device.ID
	downtimeToSave.DowntimeId = noReasonDowntime.ID
	downtimeToSave.WorkplaceId = deviceWorkplaceRecord.WorkplaceID
	db.Save(&downtimeToSave)
	LogInfo(device.Name, "Downtime opened, elapsed: "+time.Since(timer).String())

}

func OpenOrder(device zapsi_database.Device, actualWorkplaceState zapsi_database.StateRecord) {
	LogInfo(device.Name, "Opening order")
	timer := time.Now()
	connectionString, dialect := zapsi_database.CheckDatabaseType(DatabaseType, DatabaseIpAddress, DatabasePort, DatabaseLogin, DatabaseName, DatabasePassword)
	db, err := gorm.Open(dialect, connectionString)
	if err != nil {
		LogError(device.Name, "Problem opening "+DatabaseName+" database: "+err.Error())
		activeDevices = nil
		return
	}
	defer db.Close()
	var deviceWorkplaceRecord zapsi_database.DeviceWorkplaceRecord
	var order zapsi_database.Order
	db.Where("name = ?", "Internal").Find(&order)
	db.Where("device_id = ?", device.ID).Find(&deviceWorkplaceRecord)
	var orderToSave zapsi_database.OrderRecord
	orderToSave.DateTimeStart = actualWorkplaceState.DateTimeStart
	orderToSave.Duration = time.Now().Sub(actualWorkplaceState.DateTimeStart)
	orderToSave.DeviceId = device.ID
	orderToSave.WorkplaceId = deviceWorkplaceRecord.WorkplaceID
	orderToSave.OrderId = sql.NullInt32{Int32: int32(order.ID), Valid: true}
	orderToSave.Cavity = 1
	db.Save(&orderToSave)
	LogInfo(device.Name, "Order opened, elapsed: "+time.Since(timer).String())
}

func CloseDowntime(device zapsi_database.Device, openDowntimeId int) {
	LogInfo(device.Name, "Closing downtime")
	timer := time.Now()
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
	openDowntime.DateTimeEnd = sql.NullTime{Time: time.Now(), Valid: true}
	openDowntime.Duration = time.Now().Sub(openDowntime.DateTimeStart)
	db.Save(&openDowntime)
	LogInfo(device.Name, "Downtime closed, elapsed: "+time.Since(timer).String())

}

func CloseOrder(device zapsi_database.Device, openOrderId int) {
	LogInfo(device.Name, "Closing order")
	timer := time.Now()
	connectionString, dialect := zapsi_database.CheckDatabaseType(DatabaseType, DatabaseIpAddress, DatabasePort, DatabaseLogin, DatabaseName, DatabasePassword)
	db, err := gorm.Open(dialect, connectionString)
	if err != nil {
		LogError(device.Name, "Problem opening "+DatabaseName+" database: "+err.Error())
		activeDevices = nil
		return
	}
	defer db.Close()
	var deviceWorkplaceRecord zapsi_database.DeviceWorkplaceRecord
	db.Where("device_id = ?", device.ID).Find(&deviceWorkplaceRecord)
	var openOrder zapsi_database.OrderRecord
	db.Where("id=?", openOrderId).Find(&openOrder)
	var workplacePortOk zapsi_database.WorkplacePort
	db.Where("workplace_id = ?", deviceWorkplaceRecord.WorkplaceID).Where("counter_ok = ?", "1").Find(&workplacePortOk)
	countOk := 0
	db.Model(&zapsi_database.DevicePortDigitalRecord{}).Where("device_port_id = ?", workplacePortOk.DevicePortId).Where("date_time>?", openOrder.DateTimeStart).Where("data = 0").Count(&countOk)
	averageCycle := 0.0
	if countOk > 0 {
		averageCycle = time.Now().Sub(openOrder.DateTimeStart).Seconds() / float64(countOk)
	}
	openOrder.AverageCycle = float32(averageCycle)
	openOrder.DateTimeEnd = sql.NullTime{Time: time.Now(), Valid: true}
	openOrder.Duration = time.Now().Sub(openOrder.DateTimeStart)
	db.Save(&openOrder)
	LogInfo(device.Name, "Order closed, elapsed: "+time.Since(timer).String())
}

func CheckOpenDowntime(device zapsi_database.Device) int {
	LogInfo(device.Name, "Checking open downtime")
	timer := time.Now()
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
	LogInfo(device.Name, "Open downtime checked, elapsed: "+time.Since(timer).String())
	return openDowntime.ID
}

func CheckOpenOrder(device zapsi_database.Device) int {
	LogInfo(device.Name, "Checking open order")
	timer := time.Now()
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
	LogInfo(device.Name, "Open order checked, elapsed: "+time.Since(timer).String())
	return openOrder.ID
}

func GetActualState(device zapsi_database.Device) (zapsi_database.State, zapsi_database.StateRecord) {
	LogInfo(device.Name, "Downloading actual state")
	timer := time.Now()
	connectionString, dialect := zapsi_database.CheckDatabaseType(DatabaseType, DatabaseIpAddress, DatabasePort, DatabaseLogin, DatabaseName, DatabasePassword)
	db, err := gorm.Open(dialect, connectionString)
	if err != nil {
		LogError(device.Name, "Problem opening "+DatabaseName+" database: "+err.Error())
		activeDevices = nil
		return zapsi_database.State{}, zapsi_database.StateRecord{}
	}
	defer db.Close()
	var deviceWorkplaceRecord zapsi_database.DeviceWorkplaceRecord
	db.Where("device_id = ?", device.ID).Find(&deviceWorkplaceRecord)
	var workplaceState zapsi_database.StateRecord
	db.Where("workplace_id=?", deviceWorkplaceRecord.WorkplaceID).Last(&workplaceState)
	var actualState zapsi_database.State
	db.Where("id=?", workplaceState.StateId).Last(&actualState)
	LogInfo(device.Name, "Actual state downloaded, elapsed: "+time.Since(timer).String())
	return actualState, workplaceState

}
