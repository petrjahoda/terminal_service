package main

import (
	"database/sql"
	"github.com/petrjahoda/database"
	_ "gorm.io/driver/postgres"
	"gorm.io/gorm"
	"time"
)

func updateOpenOrderData(device database.Device, db *gorm.DB, openOrderRecord database.OrderRecord) {
	logInfo(device.Name, "Updating order data")
	timer := time.Now()
	var countOk int64
	var countNok int64
	var averageCycle float64
	var workplacePorts []database.WorkplacePort
	deviceWorkplaceRecordSync.Lock()
	db.Where("workplace_id = ?", cachedDeviceWorkplaceRecords[device.ID].WorkplaceID).Find(&workplacePorts)
	deviceWorkplaceRecordSync.Unlock()
	for _, port := range workplacePorts {
		if port.CounterOK {
			db.Model(&database.DevicePortDigitalRecord{}).Where("device_port_id = ?", port.DevicePortID).Where("date_time>?", openOrderRecord.DateTimeStart).Where("data = 0").Count(&countOk)
			if countOk > 0 {
				averageCycle = time.Now().Sub(openOrderRecord.DateTimeStart).Seconds() / float64(countOk)
			}
		} else if port.CounterNOK {
			db.Model(&database.DevicePortDigitalRecord{}).Where("device_port_id = ?", port.DevicePortID).Where("date_time>?", openOrderRecord.DateTimeStart).Where("data = 0").Count(&countNok)
		}
	}
	deviceWorkplaceRecordSync.Lock()
	db.Model(&openOrderRecord).Update("average_cycle", float32(averageCycle)).Update("count_ok", int(countOk)).Update("count_nok", int(countNok)).Update("workplace_id", cachedDeviceWorkplaceRecords[device.ID].WorkplaceID)
	deviceWorkplaceRecordSync.Unlock()
	logInfo(device.Name, "Order data updated in "+time.Since(timer).String())
}

func createNewDowntime(device database.Device, db *gorm.DB) {
	logInfo(device.Name, "Create new downtime")
	timer := time.Now()
	var downtimeToSave database.DowntimeRecord
	downtimeToSave.DateTimeStart = time.Now()
	downtimeToSave.DowntimeID = 1
	downtimeToSave.WorkplaceID = cachedDeviceWorkplaceRecords[device.ID].WorkplaceID
	db.Save(&downtimeToSave)
	logInfo(device.Name, "New downtime created in "+time.Since(timer).String())

}

func createNewOrder(device database.Device, db *gorm.DB, timezone string) {
	logInfo(device.Name, "Creating new order")
	timer := time.Now()
	location, err := time.LoadLocation(timezone)
	if err != nil {
		logError("MAIN", "Cannot create order, problem loading location: "+timezone)
		return
	}
	var workshiftID int
	var workplaceWorkshifts []database.WorkplaceWorkshift
	db.Where("workplace_id = ?", cachedDeviceWorkplaceRecords[device.ID].WorkplaceID).Find(&workplaceWorkshifts)
	for _, workplaceWorkshift := range workplaceWorkshifts {
		var workshift database.Workshift
		db.Where("id = ?", workplaceWorkshift.WorkshiftID).Find(&workshift)
		if workshift.WorkshiftStart.Hour() <= time.Now().In(location).Hour() && workshift.WorkshiftEnd.Hour() >= time.Now().In(location).Hour() {
			logInfo(device.Name, "Actual workshift: "+workshift.Name)
			workshiftID = int(workshift.ID)
			break
		} else if workshift.WorkshiftStart.Hour() >= workshift.WorkshiftEnd.Hour() {
			if time.Now().In(location).Hour() <= workshift.WorkshiftEnd.Hour() || time.Now().In(location).Hour() >= workshift.WorkshiftStart.Hour() {
				logInfo(device.Name, "Actual workshift: "+workshift.Name)
				workshiftID = int(workshift.ID)
				break
			}
		}
	}
	var orderToSave database.OrderRecord
	orderToSave.DateTimeStart = time.Now()
	orderToSave.WorkplaceID = cachedDeviceWorkplaceRecords[device.ID].WorkplaceID
	orderToSave.OrderID = 1
	orderToSave.WorkplaceModeID = 1
	orderToSave.WorkshiftID = workshiftID
	orderToSave.OperationID = 1
	orderToSave.Cavity = 1
	db.Save(&orderToSave)
	var userToSave database.UserRecord
	userToSave.DateTimeStart = time.Now()
	userToSave.OrderRecordID = int(orderToSave.ID)
	userToSave.UserID = 1
	userToSave.WorkplaceID = cachedDeviceWorkplaceRecords[device.ID].WorkplaceID
	db.Save(&userToSave)
	logInfo(device.Name, "New Order created in "+time.Since(timer).String())
}

func updateDowntimeToClosed(device database.Device, db *gorm.DB, openDowntimeRecord database.DowntimeRecord) {
	logInfo(device.Name, "Updating downtime to closed")
	timer := time.Now()
	db.Model(&openDowntimeRecord).Update("date_time_end", sql.NullTime{Time: time.Now(), Valid: true})
	logInfo(device.Name, "Downtime updated to closed in "+time.Since(timer).String())

}

func updateOrderToClosed(device database.Device, db *gorm.DB, openOrderRecord database.OrderRecord) {
	logInfo(device.Name, "Updating order to closed")
	timer := time.Now()
	var countOk int64
	var countNok int64
	var averageCycle float64
	var workplacePorts []database.WorkplacePort
	deviceWorkplaceRecordSync.Lock()
	db.Where("workplace_id = ?", cachedDeviceWorkplaceRecords[device.ID].WorkplaceID).Find(&workplacePorts)
	deviceWorkplaceRecordSync.Unlock()
	for _, port := range workplacePorts {
		if port.CounterOK {
			db.Model(&database.DevicePortDigitalRecord{}).Where("device_port_id = ?", port.DevicePortID).Where("date_time>?", openOrderRecord.DateTimeStart).Where("data = 0").Count(&countOk)
			if countOk > 0 {
				averageCycle = time.Now().Sub(openOrderRecord.DateTimeStart).Seconds() / float64(countOk)
			}
		} else if port.CounterNOK {
			db.Model(&database.DevicePortDigitalRecord{}).Where("device_port_id = ?", port.DevicePortID).Where("date_time>?", openOrderRecord.DateTimeStart).Where("data = 0").Count(&countNok)
		}
	}
	db.Model(&openOrderRecord).Update("average_cycle", float32(averageCycle)).Update("date_time_end", sql.NullTime{Time: time.Now(), Valid: true})

	var openUserRecord database.UserRecord
	db.Where("order_record_id = ?", openOrderRecord.ID).Find(&openUserRecord)
	db.Model(&openUserRecord).Update("date_time_end", sql.NullTime{Time: time.Now(), Valid: true})

	logInfo(device.Name, "Order updated to closed in "+time.Since(timer).String())
}

func readOpenDowntime(device database.Device, db *gorm.DB) database.DowntimeRecord {
	logInfo(device.Name, "Reading open downtime")
	timer := time.Now()
	var openDowntime database.DowntimeRecord
	deviceWorkplaceRecordSync.Lock()
	db.Where("workplace_id=?", cachedDeviceWorkplaceRecords[device.ID].WorkplaceID).Where("date_time_end is null").Last(&openDowntime)
	deviceWorkplaceRecordSync.Unlock()
	logInfo(device.Name, "Open downtime read in "+time.Since(timer).String())
	return openDowntime
}

func readOpenOrder(device database.Device, db *gorm.DB) database.OrderRecord {
	logInfo(device.Name, "Reading open order")
	timer := time.Now()
	var openOrder database.OrderRecord
	deviceWorkplaceRecordSync.Lock()
	db.Where("workplace_id=?", cachedDeviceWorkplaceRecords[device.ID].WorkplaceID).Where("date_time_end is null").Last(&openOrder)
	deviceWorkplaceRecordSync.Unlock()
	logInfo(device.Name, "Open order read in "+time.Since(timer).String())
	return openOrder
}

func readActualState(device database.Device, db *gorm.DB) database.State {
	logInfo(device.Name, "Reading actual state")
	timer := time.Now()
	var workplaceState database.StateRecord
	deviceWorkplaceRecordSync.Lock()
	db.Where("workplace_id=?", cachedDeviceWorkplaceRecords[device.ID].WorkplaceID).Last(&workplaceState)
	deviceWorkplaceRecordSync.Unlock()
	logInfo(device.Name, "Actual state read in "+time.Since(timer).String())
	stateSync.Lock()
	state := cachedStates[uint(workplaceState.StateID)]
	stateSync.Unlock()
	return state

}
