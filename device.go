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

func UpdateOrderData(device database.Device, deviceOrderRecordId int) {
	LogInfo(device.Name, "Updating order data")
	timer := time.Now()
	db, err := gorm.Open(postgres.Open(config), &gorm.Config{})
	if err != nil {
		LogError(device.Name, "Problem opening database: "+err.Error())
		activeDevices = nil
		return
	}
	sqlDB, err := db.DB()
	defer sqlDB.Close()
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
	LogInfo(device.Name, "Order data updated, elapsed: "+time.Since(timer).String())

}

func OpenDowntime(device database.Device) {
	LogInfo(device.Name, "Opening downtime")
	timer := time.Now()
	db, err := gorm.Open(postgres.Open(config), &gorm.Config{})
	if err != nil {
		LogError(device.Name, "Problem opening database: "+err.Error())
		activeDevices = nil
		return
	}
	sqlDB, err := db.DB()
	defer sqlDB.Close()
	var noReasonDowntime database.Downtime
	db.Where("name = ?", "No reason Downtime").Find(&noReasonDowntime)
	var deviceWorkplaceRecord database.DeviceWorkplaceRecord
	db.Where("device_id = ?", device.ID).Find(&deviceWorkplaceRecord)
	var downtimeToSave database.DownTimeRecord
	downtimeToSave.DateTimeStart = time.Now()
	downtimeToSave.DowntimeID = int(noReasonDowntime.ID)
	downtimeToSave.WorkplaceID = deviceWorkplaceRecord.WorkplaceID
	db.Save(&downtimeToSave)
	LogInfo(device.Name, "Downtime opened, elapsed: "+time.Since(timer).String())

}

func OpenOrder(device database.Device, timezone string) {
	LogInfo(device.Name, "Opening order")
	timer := time.Now()
	location, err := time.LoadLocation(timezone)
	if err != nil {
		LogError("MAIN", "Cannot start order, problem loading location: "+timezone)
		return
	}
	db, err := gorm.Open(postgres.Open(config), &gorm.Config{})
	if err != nil {
		LogError(device.Name, "Problem opening database: "+err.Error())
		activeDevices = nil
		return
	}
	sqlDB, err := db.DB()
	defer sqlDB.Close()
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
			LogInfo(device.Name, "Actual workshift: "+workshift.Name)
			workshiftID = int(workshift.ID)
			break
		} else if workshift.WorkshiftStart.Hour() > workshift.WorkshiftEnd.Hour() {
			if time.Now().In(location).Hour() < workshift.WorkshiftEnd.Hour() || time.Now().In(location).Hour() > workshift.WorkshiftStart.Hour() {
				LogInfo(device.Name, "Actual workshift: "+workshift.Name)
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
	db.Save(&userToSave)

	LogInfo(device.Name, "Order opened, elapsed: "+time.Since(timer).String())
}

func CloseDowntime(device database.Device, openDowntimeId int) {
	LogInfo(device.Name, "Closing downtime")
	timer := time.Now()
	db, err := gorm.Open(postgres.Open(config), &gorm.Config{})
	if err != nil {
		LogError(device.Name, "Problem opening database: "+err.Error())
		activeDevices = nil
		return
	}
	sqlDB, err := db.DB()
	defer sqlDB.Close()
	var openDowntime database.DownTimeRecord
	db.Where("id=?", openDowntimeId).Find(&openDowntime)
	openDowntime.DateTimeEnd = sql.NullTime{Time: time.Now(), Valid: true}
	db.Save(&openDowntime)
	LogInfo(device.Name, "Downtime closed, elapsed: "+time.Since(timer).String())

}

func CloseOrder(device database.Device, openOrderId int) {
	LogInfo(device.Name, "Closing order")
	timer := time.Now()
	db, err := gorm.Open(postgres.Open(config), &gorm.Config{})
	if err != nil {
		LogError(device.Name, "Problem opening  database: "+err.Error())
		activeDevices = nil
		return
	}
	sqlDB, err := db.DB()
	defer sqlDB.Close()
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

	LogInfo(device.Name, "Order closed, elapsed: "+time.Since(timer).String())
}

func CheckOpenDowntime(device database.Device) int {
	LogInfo(device.Name, "Checking open downtime")
	timer := time.Now()
	db, err := gorm.Open(postgres.Open(config), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		LogError(device.Name, "Problem opening database: "+err.Error())
		activeDevices = nil
		return 0
	}
	sqlDB, err := db.DB()
	defer sqlDB.Close()
	var openDowntime database.DownTimeRecord
	db.Where("device_id=?", device.ID).Where("date_time_end is null").Last(&openDowntime)
	LogInfo(device.Name, "Open downtime checked, elapsed: "+time.Since(timer).String())
	return int(openDowntime.ID)
}

func CheckOpenOrder(device database.Device) int {
	LogInfo(device.Name, "Checking open order")
	timer := time.Now()
	db, err := gorm.Open(postgres.Open(config), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		LogError(device.Name, "Problem opening database: "+err.Error())
		activeDevices = nil
		return 0
	}
	sqlDB, err := db.DB()
	defer sqlDB.Close()
	var openOrder database.OrderRecord
	db.Where("device_id=?", device.ID).Where("date_time_end is null").Last(&openOrder)
	LogInfo(device.Name, "Open order checked, elapsed: "+time.Since(timer).String())
	return int(openOrder.ID)
}

func GetActualState(device database.Device) database.State {
	LogInfo(device.Name, "Downloading actual state")
	timer := time.Now()
	db, err := gorm.Open(postgres.Open(config), &gorm.Config{})
	if err != nil {
		LogError(device.Name, "Problem opening database: "+err.Error())
		activeDevices = nil
		return database.State{}
	}
	sqlDB, err := db.DB()
	defer sqlDB.Close()
	var deviceWorkplaceRecord database.DeviceWorkplaceRecord
	db.Where("device_id = ?", device.ID).Find(&deviceWorkplaceRecord)
	var workplaceState database.StateRecord
	db.Where("workplace_id=?", deviceWorkplaceRecord.WorkplaceID).Last(&workplaceState)
	var actualState database.State
	db.Where("id=?", workplaceState.StateID).Last(&actualState)
	LogInfo(device.Name, "Actual state downloaded, elapsed: "+time.Since(timer).String())
	return actualState

}
