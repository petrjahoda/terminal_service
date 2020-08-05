package main

import (
	"github.com/TwinProduction/go-color"
	"github.com/petrjahoda/database"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"strconv"
	"time"
)

func UpdateProgramVersion() {
	LogInfo("MAIN", "Writing program version into settings")
	timer := time.Now()
	db, err := gorm.Open(postgres.Open(config), &gorm.Config{})
	if err != nil {
		LogError("MAIN", "Problem opening database: "+err.Error())
		return
	}
	sqlDB, err := db.DB()
	defer sqlDB.Close()
	var existingSettings database.Setting
	db.Where("name=?", serviceName).Find(&existingSettings)
	existingSettings.Name = serviceName
	existingSettings.Value = version
	db.Save(&existingSettings)
	LogInfo("MAIN", "Program version written into settings in "+time.Since(timer).String())
}

func CheckDeviceInRunningDevices(device database.Device) bool {
	for _, runningDevice := range runningDevices {
		if runningDevice.Name == device.Name {
			return true
		}
	}
	return false
}

func RunDevice(device database.Device) {
	LogInfo(device.Name, "Device active, started running")
	deviceSync.Lock()
	runningDevices = append(runningDevices, device)
	deviceSync.Unlock()
	deviceIsActive := true
	timezone := ReadTimeZoneFromDatabase()
	for deviceIsActive && serviceRunning {
		LogInfo(device.Name, "Device main loop started")
		timer := time.Now()
		actualState := ReadActualState(device)
		var stateNameColored string
		if actualState.Name == "Poweroff" {
			stateNameColored = color.Ize(color.Red, actualState.Name)
		} else if actualState.Name == "Downtime" {
			stateNameColored = color.Ize(color.Yellow, actualState.Name)
		} else {
			stateNameColored = color.Ize(color.White, actualState.Name)
		}
		LogInfo(device.Name, "Actual workplace state: "+stateNameColored)
		openOrderId := ReadOpenOrder(device)
		LogInfo(device.Name, "Actual open order: "+strconv.Itoa(openOrderId))
		openDowntimeId := ReadOpenDowntime(device)
		LogInfo(device.Name, "Actual open downtime: "+strconv.Itoa(openDowntimeId))
		orderIsOpen := openOrderId > 0
		downtimeIsOpen := openDowntimeId > 0
		switch actualState.Name {
		case "Poweroff":
			{
				LogInfo(device.Name, "Poweroff state")
				if orderIsOpen {
					UpdateOrderToClosed(device, openOrderId)
				}
				if downtimeIsOpen {
					UpdateDowntimeToClosed(device, openDowntimeId)
				}
			}
		case "Production":
			{
				LogInfo(device.Name, "Production state")
				if !orderIsOpen {
					CreateNewOrder(device, timezone)
				}
				if downtimeIsOpen {
					UpdateDowntimeToClosed(device, openDowntimeId)
				}
			}
		case "Downtime":
			{
				LogInfo(device.Name, "Downtime state")
				if !downtimeIsOpen {
					CreateNewDowntime(device)
				}
			}
		}
		if orderIsOpen {
			UpdateOpenOrderData(device, openOrderId)
		}

		LogInfo(device.Name, "Device main loop ended in "+time.Since(timer).String())
		Sleep(device, timer)
		deviceIsActive = CheckActive(device)
	}
	RemoveDeviceFromRunningDevices(device)
	LogInfo(device.Name, "Device not active, stopped running")

}

func Sleep(device database.Device, start time.Time) {
	if time.Since(start) < (downloadInSeconds * time.Second) {
		sleepTime := downloadInSeconds*time.Second - time.Since(start)
		LogInfo(device.Name, "Sleeping for "+sleepTime.String())
		time.Sleep(sleepTime)
	}
}

func CheckActive(device database.Device) bool {
	for _, activeDevice := range activeDevices {
		if activeDevice.Name == device.Name {
			LogInfo(device.Name, "Device still active")
			return true
		}
	}
	LogInfo(device.Name, "Device not active")
	return false
}

func RemoveDeviceFromRunningDevices(device database.Device) {
	deviceSync.Lock()
	for idx, runningDevice := range runningDevices {
		if device.Name == runningDevice.Name {
			runningDevices = append(runningDevices[0:idx], runningDevices[idx+1:]...)
		}
	}
	deviceSync.Unlock()
}

func ReadActiveDevices(reference string) {
	LogInfo("MAIN", "Reading active devices")
	timer := time.Now()
	db, err := gorm.Open(postgres.Open(config), &gorm.Config{})
	if err != nil {
		LogError(reference, "Problem opening database: "+err.Error())
		activeDevices = nil
		return
	}
	sqlDB, err := db.DB()
	defer sqlDB.Close()
	var deviceType database.DeviceType
	db.Where("name=?", "Zapsi Touch").Find(&deviceType)
	db.Where("device_type_id=?", deviceType.ID).Where("activated = ?", "1").Find(&activeDevices)
	LogInfo("MAIN", "Active devices read in "+time.Since(timer).String())
}

func ReadTimeZoneFromDatabase() string {
	LogInfo("MAIN", "Reading timezone from database")
	timer := time.Now()
	db, err := gorm.Open(postgres.Open(config), &gorm.Config{})
	if err != nil {
		LogError("MAIN", "Problem opening database: "+err.Error())
		return ""
	}
	sqlDB, err := db.DB()
	defer sqlDB.Close()
	var settings database.Setting
	db.Where("name=?", "timezone").Find(&settings)
	LogInfo("MAIN", "Timezone read in "+time.Since(timer).String())
	return settings.Value
}
