package main

import (
	"github.com/jinzhu/gorm"
	"github.com/petrjahoda/zapsi_database"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"time"
)

const version = "2020.1.2.1"
const programName = "Terminal Service"
const deleteLogsAfter = 240 * time.Hour
const downloadInSeconds = 10

var (
	activeDevices  []zapsi_database.Device
	runningDevices []zapsi_database.Device
	deviceSync     sync.Mutex
)

func main() {
	LogDirectoryFileCheck("MAIN")
	LogInfo("MAIN", programName+" version "+version+" started")
	CreateConfigIfNotExists()
	LoadSettingsFromConfigFile()
	LogDebug("MAIN", "Using ["+DatabaseType+"] on "+DatabaseIpAddress+":"+DatabasePort+" with database "+DatabaseName)
	CompleteDatabaseCheck()
	for {
		start := time.Now()
		LogInfo("MAIN", "Program running")
		UpdateActiveDevices("MAIN")
		DeleteOldLogFiles()
		LogInfo("MAIN", "Active devices: "+strconv.Itoa(len(activeDevices))+", running devices: "+strconv.Itoa(len(runningDevices)))
		for _, activeDevice := range activeDevices {
			activeDeviceIsRunning := CheckDevice(activeDevice)
			if !activeDeviceIsRunning {
				go RunDevice(activeDevice)
			}
		}
		if time.Since(start) < (downloadInSeconds * time.Second) {
			sleeptime := downloadInSeconds*time.Second - time.Since(start)
			LogInfo("MAIN", "Sleeping for "+sleeptime.String())
			time.Sleep(sleeptime)
		}
	}
}

func CompleteDatabaseCheck() {
	firstRunCheckComplete := false
	for firstRunCheckComplete == false {
		databaseOk := zapsi_database.CheckDatabase(DatabaseType, DatabaseIpAddress, DatabasePort, DatabaseLogin, DatabaseName, DatabasePassword)
		tablesOk, err := zapsi_database.CheckTables(DatabaseType, DatabaseIpAddress, DatabasePort, DatabaseLogin, DatabaseName, DatabasePassword)
		if err != nil {
			LogInfo("MAIN", "Problem creating tables: "+err.Error())
		}
		if databaseOk && tablesOk {
			WriteProgramVersionIntoSettings()
			firstRunCheckComplete = true
		}
	}
}

func WriteProgramVersionIntoSettings() {
	connectionString, dialect := zapsi_database.CheckDatabaseType(DatabaseType, DatabaseIpAddress, DatabasePort, DatabaseLogin, DatabaseName, DatabasePassword)
	db, err := gorm.Open(dialect, connectionString)
	if err != nil {
		LogError("MAIN", "Problem opening "+DatabaseName+" database: "+err.Error())
		activeDevices = nil
		return
	}
	defer db.Close()
	var settings zapsi_database.Setting
	db.Where("key=?", programName).Find(&settings)
	settings.Key = programName
	settings.Value = version
	db.Save(&settings)
	LogDebug("MAIN", "Updated version in database for "+programName)
}

func CheckDevice(device zapsi_database.Device) bool {
	for _, runningDevice := range runningDevices {
		if runningDevice.Name == device.Name {
			return true
		}
	}
	return false
}

func RunDevice(device zapsi_database.Device) {
	LogInfo(device.Name, "Device started running")
	deviceSync.Lock()
	runningDevices = append(runningDevices, device)
	deviceSync.Unlock()
	deviceIsActive := true
	CreateDirectoryIfNotExists(device)
	for deviceIsActive {
		start := time.Now()
		actualState := GetActualState(device)
		LogInfo(device.Name, "Actual workplace state: "+actualState.Name)
		openOrderId := CheckOpenOrder(device)
		LogInfo(device.Name, "Actual open order: "+strconv.Itoa(int(openOrderId)))
		openDowntimeId := CheckOpenDowntime(device)
		LogInfo(device.Name, "Actual open downtime: "+strconv.Itoa(int(openDowntimeId)))
		orderIsOpen := openOrderId > 0
		downtimeIsOpen := openDowntimeId > 0
		switch actualState.Name {
		case "PowerOff":
			{
				if orderIsOpen {
					closeOrder(actualState)
				}
				if downtimeIsOpen {
					closeDowntime(actualState)
				}
			}
		case "Production":
			{
				if !orderIsOpen {
					openOrder(actualState)
				}
				if downtimeIsOpen {
					closeDowntime(actualState)
				}
			}
		case "Downtime":
			{
				if !downtimeIsOpen {
					openDowntime(actualState)
				}
			}
		}
		if orderIsOpen {
			updateOrderData(actualState)
		}

		LogInfo(device.Name, "Processing takes "+time.Since(start).String())
		Sleep(device, start)
		deviceIsActive = CheckActive(device)
	}
	RemoveDeviceFromRunningDevices(device)
	LogInfo(device.Name, "Device not active, stopped running")

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
	var openorder zapsi_database.DeviceOrderRecord
	db.Where("device_id=?", device.ID).Where("date_time_end is null").Last(&openorder)
	return openorder.ID
}

func GetActualState(device zapsi_database.Device) zapsi_database.State {
	connectionString, dialect := zapsi_database.CheckDatabaseType(DatabaseType, DatabaseIpAddress, DatabasePort, DatabaseLogin, DatabaseName, DatabasePassword)
	db, err := gorm.Open(dialect, connectionString)
	if err != nil {
		LogError(device.Name, "Problem opening "+DatabaseName+" database: "+err.Error())
		activeDevices = nil
		return zapsi_database.State{}
	}
	defer db.Close()
	var workplaceState zapsi_database.WorkplaceState
	db.Where("workplace_id=?", device.WorkplaceId).Last(&workplaceState)
	var actualState zapsi_database.State
	db.Where("id=?", workplaceState.Id).Last(&actualState)
	return actualState

}

func CreateDirectoryIfNotExists(device zapsi_database.Device) {
	deviceDirectory := filepath.Join(".", strconv.Itoa(int(device.ID))+"-"+device.Name)

	if _, checkPathError := os.Stat(deviceDirectory); checkPathError == nil {
		LogInfo(device.Name, "Device directory exists")
	} else if os.IsNotExist(checkPathError) {
		LogWarning(device.Name, "Device directory not exist, creating")
		mkdirError := os.MkdirAll(deviceDirectory, 0777)
		if mkdirError != nil {
			LogError(device.Name, "Unable to create device directory: "+mkdirError.Error())
		} else {
			LogInfo(device.Name, "Device directory created")
		}
	} else {
		LogError(device.Name, "Device directory does not exist")
	}
}

func Sleep(device zapsi_database.Device, start time.Time) {
	if time.Since(start) < (downloadInSeconds * time.Second) {
		sleepTime := downloadInSeconds*time.Second - time.Since(start)
		LogInfo(device.Name, "Sleeping for "+sleepTime.String())
		time.Sleep(sleepTime)
	}
}

func CheckActive(device zapsi_database.Device) bool {
	for _, activeDevice := range activeDevices {
		if activeDevice.Name == device.Name {
			LogInfo(device.Name, "Device still active")
			return true
		}
	}
	LogInfo(device.Name, "Device not active")
	return false
}

func RemoveDeviceFromRunningDevices(device zapsi_database.Device) {
	for idx, runningDevice := range runningDevices {
		if device.Name == runningDevice.Name {
			deviceSync.Lock()
			runningDevices = append(runningDevices[0:idx], runningDevices[idx+1:]...)
			deviceSync.Unlock()
		}
	}
}

func UpdateActiveDevices(reference string) {
	connectionString, dialect := zapsi_database.CheckDatabaseType(DatabaseType, DatabaseIpAddress, DatabasePort, DatabaseLogin, DatabaseName, DatabasePassword)
	db, err := gorm.Open(dialect, connectionString)
	if err != nil {
		LogError(reference, "Problem opening "+DatabaseName+" database: "+err.Error())
		activeDevices = nil
		return
	}
	defer db.Close()
	var deviceType zapsi_database.DeviceType
	db.Where("name=?", "Zapsi Touch").Find(&deviceType)
	db.Where("device_type_id=?", deviceType.ID).Where("activated = true").Where("workplace !=?", 0).Find(&activeDevices)
	LogDebug("MAIN", "Zapsi touch device type id is "+strconv.Itoa(int(deviceType.ID)))
}
