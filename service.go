package main

import (
	"github.com/TwinProduction/go-color"
	"github.com/petrjahoda/database"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"strconv"
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
		actualState, actualStateRecord := readActualState(device, db)
		openOrderRecord := readOpenOrder(device, db)
		logInfo(device.Name, "Actual open order: "+strconv.Itoa(openOrderRecord.OrderID))
		openDowntimeRecord := readOpenDowntime(device, db)
		logInfo(device.Name, "Actual open downtime: "+strconv.Itoa(openDowntimeRecord.DowntimeID))
		orderIsOpen := openOrderRecord.ID > 0
		downtimeIsOpen := openDowntimeRecord.ID > 0
		switch actualState.Name {
		case "Poweroff":
			{
				logInfo(device.Name, color.Ize(color.Red, actualState.Name+" state"))
				if orderIsOpen {
					updateOrderToClosed(device, db, openOrderRecord, actualStateRecord)
				}
				if downtimeIsOpen {
					updateDowntimeToClosed(device, db, openDowntimeRecord, actualStateRecord)
				}
			}
		case "Production":
			{
				logInfo(device.Name, color.Ize(color.White, actualState.Name+" state"))
				if !orderIsOpen {
					createNewOrder(device, db, timezone, actualStateRecord)
				}
				if downtimeIsOpen {
					updateDowntimeToClosed(device, db, openDowntimeRecord, actualStateRecord)
				}
			}
		case "Downtime":
			{
				logInfo(device.Name, color.Ize(color.Yellow, actualState.Name+" state"))
				if !downtimeIsOpen {
					createNewDowntime(device, db, actualStateRecord, openOrderRecord)
				}
			}
		}
		if orderIsOpen {
			updateOpenOrderData(device, db, openOrderRecord)
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
	var deviceType database.DeviceType
	db.Where("name=?", "Zapsi Touch").Find(&deviceType)
	db.Where("device_type_id=?", deviceType.ID).Where("activated = ?", "1").Find(&activeDevices)
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
