package main

import (
	"github.com/jinzhu/gorm"
	"github.com/kardianos/service"
	"github.com/petrjahoda/zapsi_database"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"time"
)

const version = "2020.1.2.29"
const programName = "Terminal Service"
const programDesription = "Created default data for terminals"
const deleteLogsAfter = 240 * time.Hour
const downloadInSeconds = 10

var serviceRunning = false

var (
	activeDevices  []zapsi_database.Device
	runningDevices []zapsi_database.Device
	deviceSync     sync.Mutex
)

type program struct{}

func (p *program) Start(s service.Service) error {
	LogInfo("MAIN", "Starting "+programName+" on "+s.Platform())
	go p.run()
	serviceRunning = true
	return nil
}

func (p *program) run() {
	time.Sleep(time.Second * 5)
	LogDirectoryFileCheck("MAIN")
	LogInfo("MAIN", programName+" version "+version+" started")
	CreateConfigIfNotExists()
	LoadSettingsFromConfigFile()
	LogDebug("MAIN", "Using ["+DatabaseType+"] on "+DatabaseIpAddress+":"+DatabasePort+" with database "+DatabaseName)
	WriteProgramVersionIntoSettings()
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
func (p *program) Stop(s service.Service) error {
	serviceRunning = false
	for len(runningDevices) != 0 {
		LogInfo("MAIN", "Stopping, still running devices: "+strconv.Itoa(len(runningDevices)))
		time.Sleep(1 * time.Second)
	}
	LogInfo("MAIN", "Stopped on platform "+s.Platform())
	return nil
}
func main() {
	serviceConfig := &service.Config{
		Name:        programName,
		DisplayName: programName,
		Description: programDesription,
	}
	prg := &program{}
	s, err := service.New(prg, serviceConfig)
	if err != nil {
		LogError("MAIN", err.Error())
	}
	err = s.Run()
	if err != nil {
		LogError("MAIN", "Problem starting "+serviceConfig.Name)
	}
}

func WriteProgramVersionIntoSettings() {
	LogInfo("MAIN", "Updating program version in database")
	timer := time.Now()
	connectionString, dialect := zapsi_database.CheckDatabaseType(DatabaseType, DatabaseIpAddress, DatabasePort, DatabaseLogin, DatabaseName, DatabasePassword)
	db, err := gorm.Open(dialect, connectionString)
	if err != nil {
		LogError("MAIN", "Problem opening "+DatabaseName+" database: "+err.Error())
		activeDevices = nil
		return
	}
	defer db.Close()
	var settings zapsi_database.Setting
	db.Where("name=?", programName).Find(&settings)
	settings.Name = programName
	settings.Value = version
	db.Save(&settings)
	LogInfo("MAIN", "Program version updated, elapsed: "+time.Since(timer).String())
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
	for deviceIsActive && serviceRunning {
		LogInfo(device.Name, "Starting device loop")
		timer := time.Now()
		actualState, actualWorkplaceState := GetActualState(device)
		LogInfo(device.Name, "Actual workplace state: "+actualState.Name)
		openOrderId := CheckOpenOrder(device)
		LogInfo(device.Name, "Actual open order: "+strconv.Itoa(openOrderId))
		openDowntimeId := CheckOpenDowntime(device)
		LogInfo(device.Name, "Actual open downtime: "+strconv.Itoa(openDowntimeId))
		orderIsOpen := openOrderId > 0
		downtimeIsOpen := openDowntimeId > 0
		switch actualState.Name {
		case "Poweroff":
			{
				LogInfo(device.Name, "Poweroff state")
				if orderIsOpen {
					CloseOrder(device, openOrderId)
				}
				if downtimeIsOpen {
					CloseDowntime(device, openDowntimeId)
				}
			}
		case "Production":
			{
				LogInfo(device.Name, "Production state")
				if !orderIsOpen {
					OpenOrder(device, actualWorkplaceState)
				}
				if downtimeIsOpen {
					CloseDowntime(device, openDowntimeId)
				}
			}
		case "Downtime":
			{
				LogInfo(device.Name, "Downtime state")
				if !downtimeIsOpen {
					OpenDowntime(device, actualWorkplaceState, openOrderId)
				}
			}
		}
		if downtimeIsOpen {
			UpdateDowntimeData(device, openDowntimeId)
		}
		if orderIsOpen {
			UpdateOrderData(device, openOrderId)
		}

		LogInfo(device.Name, "Loop ended, elapsed: "+time.Since(timer).String())
		Sleep(device, timer)
		deviceIsActive = CheckActive(device)
	}
	RemoveDeviceFromRunningDevices(device)
	LogInfo(device.Name, "Device not active, stopped running")

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
	deviceSync.Lock()
	for idx, runningDevice := range runningDevices {
		if device.Name == runningDevice.Name {
			runningDevices = append(runningDevices[0:idx], runningDevices[idx+1:]...)
		}
	}
	deviceSync.Unlock()
}

func UpdateActiveDevices(reference string) {
	LogInfo("MAIN", "Updating active devices")
	timer := time.Now()
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
	db.Where("device_type_id=?", deviceType.ID).Where("activated = ?", "1").Find(&activeDevices)
	LogInfo("MAIN", "Active devices updated, elapsed: "+time.Since(timer).String())
}
