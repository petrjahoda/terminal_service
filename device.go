package main

import (
	"database/sql"
	"github.com/petrjahoda/zapsi_database"
	"gorm.io/driver/postgres"
	_ "gorm.io/driver/postgres"
	"gorm.io/gorm"
	"time"
)

func UpdateOrderData(device zapsi_database.Device, deviceOrderRecordId int) {
	LogInfo(device.Name, "Updating order data")
	timer := time.Now()
	db, err := gorm.Open(postgres.Open(config), &gorm.Config{})
	if err != nil {
		LogError(device.Name, "Problem opening database: "+err.Error())
		activeDevices = nil
		return
	}
	var deviceWorkplaceRecord zapsi_database.DeviceWorkplaceRecord
	db.Where("device_id = ?", device.ID).Find(&deviceWorkplaceRecord)
	var openOrder zapsi_database.OrderRecord
	db.Where("id=?", deviceOrderRecordId).Find(&openOrder)
	var workplacePortOk zapsi_database.WorkplacePort
	db.Where("workplace_id = ?", deviceWorkplaceRecord.WorkplaceID).Where("counter_ok = ?", "1").Find(&workplacePortOk)
	var countOk int64
	db.Model(&zapsi_database.DevicePortDigitalRecord{}).Where("device_port_id = ?", workplacePortOk.DevicePortID).Where("date_time>?", openOrder.DateTimeStart).Where("data = 0").Count(&countOk)
	var workplacePortNok zapsi_database.WorkplacePort
	db.Where("workplace_id = ?", deviceWorkplaceRecord.WorkplaceID).Where("counter_nok = ?", "1").Find(&workplacePortNok)
	var countNok int64
	db.Model(&zapsi_database.DevicePortDigitalRecord{}).Where("device_port_id = ?", workplacePortNok.DevicePortID).Where("date_time>?", openOrder.DateTimeStart).Where("data = 0").Count(&countNok)
	averageCycle := 0.0
	if countOk > 0 {
		averageCycle = time.Now().Sub(openOrder.DateTimeStart).Seconds() / float64(countOk)
	}
	openOrder.AverageCycle = float32(averageCycle)
	openOrder.CountOk = int(countOk)
	openOrder.CountNok = int(countNok)
	openOrder.WorkplaceID = deviceWorkplaceRecord.WorkplaceID
	db.Save(&openOrder)
	LogInfo(device.Name, "Order data updated, elapsed: "+time.Since(timer).String())

}

func OpenDowntime(device zapsi_database.Device, actualWorkplaceState zapsi_database.StateRecord, openOrderId int) {
	LogInfo(device.Name, "Opening downtime")
	timer := time.Now()
	db, err := gorm.Open(postgres.Open(config), &gorm.Config{})
	if err != nil {
		LogError(device.Name, "Problem opening database: "+err.Error())
		activeDevices = nil
		return
	}
	var noReasonDowntime zapsi_database.Downtime
	db.Where("name = ?", "No reason downtime").Find(&noReasonDowntime)
	var deviceWorkplaceRecord zapsi_database.DeviceWorkplaceRecord
	db.Where("device_id = ?", device.ID).Find(&deviceWorkplaceRecord)
	var downtimeToSave zapsi_database.DownTimeRecord
	downtimeToSave.DateTimeStart = actualWorkplaceState.DateTimeStart
	if openOrderId > 0 {
		downtimeToSave.OrderRecordID = openOrderId
		var deviceUserRecord zapsi_database.UserRecord
		db.Where("order_record_id = ?", openOrderId).Find(&deviceUserRecord)
		userIsValid := deviceUserRecord.UserID != 0
		if userIsValid {
			downtimeToSave.UserID = deviceUserRecord.UserID
		}
	}
	downtimeToSave.DeviceID = int(device.ID)
	downtimeToSave.DowntimeID = int(noReasonDowntime.ID)
	downtimeToSave.WorkplaceID = deviceWorkplaceRecord.WorkplaceID
	db.Save(&downtimeToSave)
	LogInfo(device.Name, "Downtime opened, elapsed: "+time.Since(timer).String())

}

func OpenOrder(device zapsi_database.Device, actualWorkplaceState zapsi_database.StateRecord) {
	LogInfo(device.Name, "Opening order")
	timer := time.Now()
	db, err := gorm.Open(postgres.Open(config), &gorm.Config{})
	if err != nil {
		LogError(device.Name, "Problem opening database: "+err.Error())
		activeDevices = nil
		return
	}
	var deviceWorkplaceRecord zapsi_database.DeviceWorkplaceRecord
	var order zapsi_database.Order
	db.Where("name = ?", "Internal").Find(&order)
	db.Where("device_id = ?", device.ID).Find(&deviceWorkplaceRecord)
	var orderToSave zapsi_database.OrderRecord
	orderToSave.DateTimeStart = actualWorkplaceState.DateTimeStart
	orderToSave.DeviceID = int(device.ID)
	orderToSave.WorkplaceID = deviceWorkplaceRecord.WorkplaceID
	orderToSave.OrderID = int(order.ID)
	orderToSave.Cavity = 1
	db.Save(&orderToSave)
	LogInfo(device.Name, "Order opened, elapsed: "+time.Since(timer).String())
}

func CloseDowntime(device zapsi_database.Device, openDowntimeId int) {
	LogInfo(device.Name, "Closing downtime")
	timer := time.Now()
	db, err := gorm.Open(postgres.Open(config), &gorm.Config{})
	if err != nil {
		LogError(device.Name, "Problem opening database: "+err.Error())
		activeDevices = nil
		return
	}
	var openDowntime zapsi_database.DownTimeRecord
	db.Where("id=?", openDowntimeId).Find(&openDowntime)
	openDowntime.DateTimeEnd = sql.NullTime{Time: time.Now(), Valid: true}
	db.Save(&openDowntime)
	LogInfo(device.Name, "Downtime closed, elapsed: "+time.Since(timer).String())

}

func CloseOrder(device zapsi_database.Device, openOrderId int) {
	LogInfo(device.Name, "Closing order")
	timer := time.Now()
	db, err := gorm.Open(postgres.Open(config), &gorm.Config{})
	if err != nil {
		LogError(device.Name, "Problem opening  database: "+err.Error())
		activeDevices = nil
		return
	}
	var deviceWorkplaceRecord zapsi_database.DeviceWorkplaceRecord
	db.Where("device_id = ?", device.ID).Find(&deviceWorkplaceRecord)
	var openOrder zapsi_database.OrderRecord
	db.Where("id=?", openOrderId).Find(&openOrder)
	var workplacePortOk zapsi_database.WorkplacePort
	db.Where("workplace_id = ?", deviceWorkplaceRecord.WorkplaceID).Where("counter_ok = ?", "1").Find(&workplacePortOk)
	var countOk int64
	db.Model(&zapsi_database.DevicePortDigitalRecord{}).Where("device_port_id = ?", workplacePortOk.DevicePortID).Where("date_time>?", openOrder.DateTimeStart).Where("data = 0").Count(&countOk)
	averageCycle := 0.0
	if countOk > 0 {
		averageCycle = time.Now().Sub(openOrder.DateTimeStart).Seconds() / float64(countOk)
	}
	openOrder.AverageCycle = float32(averageCycle)
	openOrder.DateTimeEnd = sql.NullTime{Time: time.Now(), Valid: true}
	db.Save(&openOrder)
	LogInfo(device.Name, "Order closed, elapsed: "+time.Since(timer).String())
}

func CheckOpenDowntime(device zapsi_database.Device) int {
	LogInfo(device.Name, "Checking open downtime")
	timer := time.Now()
	db, err := gorm.Open(postgres.Open(config), &gorm.Config{})
	if err != nil {
		LogError(device.Name, "Problem opening database: "+err.Error())
		activeDevices = nil
		return 0
	}
	var openDowntime zapsi_database.DownTimeRecord
	db.Where("device_id=?", device.ID).Where("date_time_end is null").Last(&openDowntime)
	LogInfo(device.Name, "Open downtime checked, elapsed: "+time.Since(timer).String())
	return int(openDowntime.ID)
}

func CheckOpenOrder(device zapsi_database.Device) int {
	LogInfo(device.Name, "Checking open order")
	timer := time.Now()
	db, err := gorm.Open(postgres.Open(config), &gorm.Config{})
	if err != nil {
		LogError(device.Name, "Problem opening database: "+err.Error())
		activeDevices = nil
		return 0
	}
	var openOrder zapsi_database.OrderRecord
	db.Where("device_id=?", device.ID).Where("date_time_end is null").Last(&openOrder)
	LogInfo(device.Name, "Open order checked, elapsed: "+time.Since(timer).String())
	return int(openOrder.ID)
}

func GetActualState(device zapsi_database.Device) (zapsi_database.State, zapsi_database.StateRecord) {
	LogInfo(device.Name, "Downloading actual state")
	timer := time.Now()
	db, err := gorm.Open(postgres.Open(config), &gorm.Config{})
	if err != nil {
		LogError(device.Name, "Problem opening database: "+err.Error())
		activeDevices = nil
		return zapsi_database.State{}, zapsi_database.StateRecord{}
	}
	var deviceWorkplaceRecord zapsi_database.DeviceWorkplaceRecord
	db.Where("device_id = ?", device.ID).Find(&deviceWorkplaceRecord)
	var workplaceState zapsi_database.StateRecord
	db.Where("workplace_id=?", deviceWorkplaceRecord.WorkplaceID).Last(&workplaceState)
	var actualState zapsi_database.State
	db.Where("id=?", workplaceState.StateID).Last(&actualState)
	LogInfo(device.Name, "Actual state downloaded, elapsed: "+time.Since(timer).String())
	return actualState, workplaceState

}
