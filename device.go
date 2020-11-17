package main

import (
	"database/sql"
	"github.com/petrjahoda/database"
	"gorm.io/driver/postgres"
	_ "gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"time"
)

func updateOpenOrderData(device database.Device, deviceOrderRecordId int) {
	logInfo(device.Name, "Updating order data")
	timer := time.Now()
	db, err := gorm.Open(postgres.Open(config), &gorm.Config{})
	sqlDB, _ := db.DB()
	defer sqlDB.Close()
	if err != nil {
		logError(device.Name, "Problem opening database: "+err.Error())
		activeDevices = nil
		return
	}
	var deviceWorkplaceRecord database.DeviceWorkplaceRecord
	db.Where("device_id = ?", device.ID).Find(&deviceWorkplaceRecord)
	var openOrder database.OrderRecord
	db.Where("id=?", deviceOrderRecordId).Find(&openOrder)
	var workplacePortOk database.WorkplacePort
	db.Where("workplace_id = ?", deviceWorkplaceRecord.WorkplaceID).Where("counter_ok = ?", "1").Find(&workplacePortOk)
	var countOk int64
	db.Model(&database.DevicePortDigitalRecord{}).Where("device_port_id = ?", workplacePortOk.DevicePortID).Where("date_time>?", openOrder.DateTimeStart).Where("data = 0").Count(&countOk)
	var workplacePortNok database.WorkplacePort
	db.Where("workplace_id = ?", deviceWorkplaceRecord.WorkplaceID).Where("counter_nok = ?", "1").Find(&workplacePortNok)
	var countNok int64
	db.Model(&database.DevicePortDigitalRecord{}).Where("device_port_id = ?", workplacePortNok.DevicePortID).Where("date_time>?", openOrder.DateTimeStart).Where("data = 0").Count(&countNok)
	averageCycle := 0.0
	if countOk > 0 {
		averageCycle = time.Now().Sub(openOrder.DateTimeStart).Seconds() / float64(countOk)
	}
	openOrder.AverageCycle = float32(averageCycle)
	openOrder.CountOk = int(countOk)
	openOrder.CountNok = int(countNok)
	openOrder.WorkplaceID = deviceWorkplaceRecord.WorkplaceID
	db.Save(&openOrder)
	logInfo(device.Name, "Order data updated in "+time.Since(timer).String())

}

func createNewDowntime(device database.Device) {
	logInfo(device.Name, "Create new downtime")
	timer := time.Now()
	db, err := gorm.Open(postgres.Open(config), &gorm.Config{})
	sqlDB, _ := db.DB()
	defer sqlDB.Close()
	if err != nil {
		logError(device.Name, "Problem opening database: "+err.Error())
		activeDevices = nil
		return
	}
	var noReasonDowntime database.Downtime
	db.Where("name = ?", "No reason Downtime").Find(&noReasonDowntime)
	var deviceWorkplaceRecord database.DeviceWorkplaceRecord
	db.Where("device_id = ?", device.ID).Find(&deviceWorkplaceRecord)
	var downtimeToSave database.DowntimeRecord
	downtimeToSave.DateTimeStart = time.Now()
	downtimeToSave.DowntimeID = int(noReasonDowntime.ID)
	downtimeToSave.WorkplaceID = deviceWorkplaceRecord.WorkplaceID
	db.Save(&downtimeToSave)
	logInfo(device.Name, "New downtime created in "+time.Since(timer).String())

}

func createNewOrder(device database.Device, timezone string) {
	logInfo(device.Name, "Creating new order")
	timer := time.Now()
	location, err := time.LoadLocation(timezone)
	if err != nil {
		logError("MAIN", "Cannot create order, problem loading location: "+timezone)
		return
	}
	db, err := gorm.Open(postgres.Open(config), &gorm.Config{})
	sqlDB, _ := db.DB()
	defer sqlDB.Close()
	if err != nil {
		logError(device.Name, "Problem opening database: "+err.Error())
		activeDevices = nil
		return
	}
	var deviceWorkplaceRecord database.DeviceWorkplaceRecord
	var order database.Order
	var workplace database.Workplace
	var workplaceWorkshifts []database.WorkplaceWorkshift
	var operation database.Operation
	var workshiftID int
	db.Where("name = ?", "Internal").Find(&order)
	db.Where("device_id = ?", device.ID).Find(&deviceWorkplaceRecord)
	db.Where("id = ?", deviceWorkplaceRecord.WorkplaceID).Find(&workplace)
	db.Where("workplace_id = ?", deviceWorkplaceRecord.WorkplaceID).Find(&workplaceWorkshifts)
	for _, workplaceWorkshift := range workplaceWorkshifts {
		var workshift database.Workshift
		db.Where("id = ?", workplaceWorkshift.WorkshiftID).Find(&workshift)
		if workshift.WorkshiftStart.Hour() < time.Now().In(location).Hour() && workshift.WorkshiftEnd.Hour() > time.Now().In(location).Hour() {
			logInfo(device.Name, "Actual workshift: "+workshift.Name)
			workshiftID = int(workshift.ID)
			break
		} else if workshift.WorkshiftStart.Hour() > workshift.WorkshiftEnd.Hour() {
			if time.Now().In(location).Hour() < workshift.WorkshiftEnd.Hour() || time.Now().In(location).Hour() > workshift.WorkshiftStart.Hour() {
				logInfo(device.Name, "Actual workshift: "+workshift.Name)
				workshiftID = int(workshift.ID)
				break
			}
		}
	}
	db.Where("order_id = ?", order.ID).Find(&operation)
	var orderToSave database.OrderRecord
	orderToSave.DateTimeStart = time.Now()
	orderToSave.WorkplaceID = deviceWorkplaceRecord.WorkplaceID
	orderToSave.OrderID = int(order.ID)
	orderToSave.WorkplaceModeID = workplace.WorkplaceModeID
	orderToSave.WorkshiftID = workshiftID
	orderToSave.OperationID = int(operation.ID)
	orderToSave.Cavity = 1
	db.Save(&orderToSave)
	var userToSave database.UserRecord
	userToSave.DateTimeStart = time.Now()
	userToSave.OrderRecordID = int(orderToSave.ID)
	userToSave.UserID = 1
	userToSave.WorkplaceID = workplace.WorkplaceModeID
	db.Save(&userToSave)
	logInfo(device.Name, "New Order created in "+time.Since(timer).String())
}

func updateDowntimeToClosed(device database.Device, openDowntimeId int) {
	logInfo(device.Name, "Updating downtime to closed")
	timer := time.Now()
	db, err := gorm.Open(postgres.Open(config), &gorm.Config{})
	sqlDB, _ := db.DB()
	defer sqlDB.Close()
	if err != nil {
		logError(device.Name, "Problem opening database: "+err.Error())
		activeDevices = nil
		return
	}
	var openDowntime database.DowntimeRecord
	db.Where("id=?", openDowntimeId).Find(&openDowntime)
	openDowntime.DateTimeEnd = sql.NullTime{Time: time.Now(), Valid: true}
	db.Save(&openDowntime)
	logInfo(device.Name, "Downtime updated to closed in "+time.Since(timer).String())

}

func updateOrderToClosed(device database.Device, openOrderId int) {
	logInfo(device.Name, "Updating order to closed")
	timer := time.Now()
	db, err := gorm.Open(postgres.Open(config), &gorm.Config{})
	sqlDB, _ := db.DB()
	defer sqlDB.Close()
	if err != nil {
		logError(device.Name, "Problem opening  database: "+err.Error())
		activeDevices = nil
		return
	}
	var deviceWorkplaceRecord database.DeviceWorkplaceRecord
	db.Where("device_id = ?", device.ID).Find(&deviceWorkplaceRecord)
	var openOrder database.OrderRecord
	var openUser database.UserRecord
	db.Where("id=?", openOrderId).Find(&openOrder)
	var workplacePortOk database.WorkplacePort
	db.Where("workplace_id = ?", deviceWorkplaceRecord.WorkplaceID).Where("counter_ok = ?", "1").Find(&workplacePortOk)
	db.Where("order_record_id = ?", openOrder.ID).Find(&openUser)
	var countOk int64
	db.Model(&database.DevicePortDigitalRecord{}).Where("device_port_id = ?", workplacePortOk.DevicePortID).Where("date_time>?", openOrder.DateTimeStart).Where("data = 0").Count(&countOk)
	averageCycle := 0.0
	if countOk > 0 {
		averageCycle = time.Now().Sub(openOrder.DateTimeStart).Seconds() / float64(countOk)
	}
	openOrder.AverageCycle = float32(averageCycle)
	openOrder.DateTimeEnd = sql.NullTime{Time: time.Now(), Valid: true}
	db.Save(&openOrder)
	openUser.DateTimeEnd = sql.NullTime{Time: time.Now(), Valid: true}
	db.Save(&openUser)
	logInfo(device.Name, "Order updated to closed in "+time.Since(timer).String())
}

func readOpenDowntime(device database.Device) int {
	logInfo(device.Name, "Reading open downtime")
	timer := time.Now()
	db, err := gorm.Open(postgres.Open(config), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	sqlDB, _ := db.DB()
	defer sqlDB.Close()
	if err != nil {
		logError(device.Name, "Problem opening database: "+err.Error())
		activeDevices = nil
		return 0
	}
	var deviceWorkplaceRecord database.DeviceWorkplaceRecord
	db.Where("device_id = ?", device.ID).Find(&deviceWorkplaceRecord)
	var openDowntime database.DowntimeRecord
	db.Where("workplace_id=?", deviceWorkplaceRecord.WorkplaceID).Where("date_time_end is null").Last(&openDowntime)
	logInfo(device.Name, "Open downtime read in "+time.Since(timer).String())
	return int(openDowntime.ID)
}

func readOpenOrder(device database.Device) int {
	logInfo(device.Name, "Reading open order")
	timer := time.Now()
	db, err := gorm.Open(postgres.Open(config), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	sqlDB, _ := db.DB()
	defer sqlDB.Close()
	if err != nil {
		logError(device.Name, "Problem opening database: "+err.Error())
		activeDevices = nil
		return 0
	}
	var deviceWorkplaceRecord database.DeviceWorkplaceRecord
	db.Where("device_id = ?", device.ID).Find(&deviceWorkplaceRecord)
	var openOrder database.OrderRecord
	db.Where("workplace_id=?", deviceWorkplaceRecord.WorkplaceID).Where("date_time_end is null").Last(&openOrder)
	logInfo(device.Name, "Open order read in "+time.Since(timer).String())
	return int(openOrder.ID)
}

func readActualState(device database.Device) database.State {
	logInfo(device.Name, "Reading actual state")
	timer := time.Now()
	db, err := gorm.Open(postgres.Open(config), &gorm.Config{})
	sqlDB, _ := db.DB()
	defer sqlDB.Close()
	if err != nil {
		logError(device.Name, "Problem opening database: "+err.Error())
		activeDevices = nil
		return database.State{}
	}
	var deviceWorkplaceRecord database.DeviceWorkplaceRecord
	db.Where("device_id = ?", device.ID).Find(&deviceWorkplaceRecord)
	var workplaceState database.StateRecord
	db.Where("workplace_id=?", deviceWorkplaceRecord.WorkplaceID).Last(&workplaceState)
	var actualState database.State
	db.Where("id=?", workplaceState.StateID).Last(&actualState)
	logInfo(device.Name, "Actual state read in "+time.Since(timer).String())
	return actualState

}
