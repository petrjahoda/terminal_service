package main

import (
	"github.com/TwinProduction/go-color"
	"github.com/petrjahoda/database"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"time"
)

func updateProgramVersion() {
	logInfo("MAIN", "Writing program version into settings")
	timer := time.Now()
	db, err := gorm.Open(postgres.Open(config), &gorm.Config{})
	sqlDB, _ := db.DB()
	defer sqlDB.Close()
	if err != nil {
		logError("MAIN", "Problem opening database: "+err.Error())
		return
	}
	var existingSettings database.Setting
	db.Where("name=?", serviceName).Find(&existingSettings)
	existingSettings.Name = serviceName
	existingSettings.Value = version
	db.Save(&existingSettings)
	logInfo("MAIN", "Program version written into settings in "+time.Since(timer).String())
}

func checkDeviceInRunningDevices(device database.Device) bool {
	for _, runningDevice := range runningDevices {
		if runningDevice.Name == device.Name {
			return true
		}
	}
	return false
}

func runDevice(device database.Device) {
	logInfo(device.Name, "Device active, started running")
	deviceSync.Lock()
	runningDevices = append(runningDevices, device)
	deviceSync.Unlock()
	deviceIsActive := true
	timezone := readTimeZoneFromDatabase()
	for deviceIsActive && serviceRunning {
		logInfo(device.Name, "Device main loop started")
		timer := time.Now()
		db, err := gorm.Open(postgres.Open(config), &gorm.Config{})
		sqlDB, _ := db.DB()
		if err != nil {
			logError(device.Name, "Problem opening database: "+err.Error())
			sleep(device, timer)
			continue
		}
		deviceWorkplaceRecordSync.Lock()
		deviceWorkplaceRecord := cachedDeviceWorkplaceRecords[device.ID]
		deviceWorkplaceRecordSync.Unlock()
		workplaceDowntimeRecordSync.Lock()
		downtimeIsOpen := cachedWorkplaceDowntimeRecords[deviceWorkplaceRecord.WorkplaceID] > 0
		workplaceDowntimeRecordSync.Unlock()
		workplaceOrderRecordSync.Lock()
		orderIsOpen := cachedWorkplaceOrderRecords[deviceWorkplaceRecord.WorkplaceID].ID > 0
		workplaceOrderRecordSync.Unlock()
		workplaceStateRecordSync.Lock()
		cachedWorkplaceState := cachedWorkplaceStateRecords[deviceWorkplaceRecord.WorkplaceID]
		workplaceStateRecordSync.Unlock()
		stateSync.Lock()
		cachedStateName := cachedStates[uint(cachedWorkplaceState)].Name
		stateSync.Unlock()
		switch cachedStateName {
		case "Poweroff":
			{
				logInfo(device.Name, color.Ize(color.Red, cachedStateName+" state"))
				if orderIsOpen {
					updateOrderToClosed(device, db, deviceWorkplaceRecord, cachedWorkplaceOrderRecords[deviceWorkplaceRecord.WorkplaceID])
				}
				if downtimeIsOpen {
					updateDowntimeToClosed(device, db, deviceWorkplaceRecord)
				}
			}
		case "Production":
			{
				logInfo(device.Name, color.Ize(color.White, cachedStateName+" state"))
				if !orderIsOpen {
					createNewOrder(device, db, timezone)
				}
				if downtimeIsOpen {
					updateDowntimeToClosed(device, db, deviceWorkplaceRecord)
				}
			}
		case "Downtime":
			{
				logInfo(device.Name, color.Ize(color.Yellow, cachedStateName+" state"))
				if !downtimeIsOpen {
					createNewDowntime(device, db)
				}
			}
		}
		if orderIsOpen {
			updateOpenOrderData(device, db, deviceWorkplaceRecord, cachedWorkplaceOrderRecords[deviceWorkplaceRecord.WorkplaceID])
		}

		logInfo(device.Name, "Device main loop ended in "+time.Since(timer).String())
		sqlDB.Close()
		sleep(device, timer)
		deviceIsActive = checkActive(device)
	}
	removeDeviceFromRunningDevices(device)
	logInfo(device.Name, "Device not active, stopped running")

}

func sleep(device database.Device, start time.Time) {
	if time.Since(start) < (downloadInSeconds * time.Second) {
		sleepTime := downloadInSeconds*time.Second - time.Since(start)
		logInfo(device.Name, "Sleeping for "+sleepTime.String())
		time.Sleep(sleepTime)
	}
}

func checkActive(device database.Device) bool {
	for _, activeDevice := range activeDevices {
		if activeDevice.Name == device.Name {
			logInfo(device.Name, "Device still active")
			return true
		}
	}
	logInfo(device.Name, "Device not active")
	return false
}

func removeDeviceFromRunningDevices(device database.Device) {
	deviceSync.Lock()
	for idx, runningDevice := range runningDevices {
		if device.Name == runningDevice.Name {
			runningDevices = append(runningDevices[0:idx], runningDevices[idx+1:]...)
		}
	}
	deviceSync.Unlock()
}

func readActiveDevices() {
	logInfo("MAIN", "Reading active devices")
	timer := time.Now()
	db, err := gorm.Open(postgres.Open(config), &gorm.Config{})
	sqlDB, _ := db.DB()
	defer sqlDB.Close()
	if err != nil {
		logError("MAIN", "Problem opening database: "+err.Error())
		activeDevices = nil
		return
	}
	db.Where("device_type_id=(select id from device_types where name = 'Zapsi Touch')").Where("activated = ?", "1").Find(&activeDevices)
	logInfo("MAIN", "Active devices read in "+time.Since(timer).String())
}

func readTimeZoneFromDatabase() string {
	logInfo("MAIN", "Reading timezone from database")
	timer := time.Now()
	db, err := gorm.Open(postgres.Open(config), &gorm.Config{})
	sqlDB, _ := db.DB()
	defer sqlDB.Close()
	if err != nil {
		logError("MAIN", "Problem opening database: "+err.Error())
		return ""
	}
	var settings database.Setting
	db.Where("name=?", "timezone").Find(&settings)
	logInfo("MAIN", "Timezone read in "+time.Since(timer).String())
	return settings.Value
}
