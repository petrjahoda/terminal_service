package main

import (
	"github.com/kardianos/service"
	"github.com/petrjahoda/database"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"strconv"
	"sync"
	"time"
)

const version = "2021.2.3.5"
const serviceName = "Terminal Service"
const serviceDescription = "Created default data for terminals"
const downloadInSeconds = 10
const config = "user=postgres password=pj79.. dbname=system host=database port=5432 sslmode=disable application_name=terminal_service"

var serviceRunning = false

var (
	activeDevices  []database.Device
	runningDevices []database.Device
	deviceSync     sync.Mutex
)

var (
	cachedStates = map[uint]database.State{}
	stateSync    sync.Mutex
)

type program struct{}

func main() {
	logInfo("MAIN", serviceName+" ["+version+"] starting...")
	serviceConfig := &service.Config{
		Name:        serviceName,
		DisplayName: serviceName,
		Description: serviceDescription,
	}
	prg := &program{}
	s, err := service.New(prg, serviceConfig)
	if err != nil {
		logError("MAIN", "Cannot start: "+err.Error())
	}
	err = s.Run()
	if err != nil {
		logError("MAIN", "Cannot start: "+err.Error())
	}
}

func (p *program) Start(service.Service) error {
	logInfo("MAIN", serviceName+" ["+version+"] started")
	go p.run()
	serviceRunning = true
	return nil
}

func (p *program) Stop(service.Service) error {
	serviceRunning = false
	for len(runningDevices) != 0 {
		logInfo("MAIN", serviceName+" ["+version+"] stopping...")
		time.Sleep(1 * time.Second)
	}
	logInfo("MAIN", serviceName+" ["+version+"] stopped")
	return nil
}

func (p *program) run() {
	updateProgramVersion()
	for {
		logInfo("MAIN", serviceName+" ["+version+"] running")
		start := time.Now()
		readActiveDevices()
		readActiveStates()
		logInfo("MAIN", "Active devices: "+strconv.Itoa(len(activeDevices))+", running devices: "+strconv.Itoa(len(runningDevices)))
		for _, activeDevice := range activeDevices {
			activeDeviceIsRunning := checkDeviceInRunningDevices(activeDevice)
			if !activeDeviceIsRunning {
				go runDevice(activeDevice)
			}
		}
		if time.Since(start) < (downloadInSeconds * time.Second) {
			sleepTime := downloadInSeconds*time.Second - time.Since(start)
			logInfo("MAIN", "Sleeping for "+sleepTime.String())
			time.Sleep(sleepTime)
		}
	}
}

func readActiveStates() {
	logInfo("MAIN", "Reading states")
	timer := time.Now()
	db, err := gorm.Open(postgres.Open(config), &gorm.Config{})
	sqlDB, _ := db.DB()
	defer sqlDB.Close()
	if err != nil {
		logError("MAIN", "Problem opening database: "+err.Error())
		return
	}
	var records []database.State
	db.Find(&records)
	if len(records) > 0 {
		stateSync.Lock()
		for _, record := range records {
			cachedStates[record.ID] = record
		}
		stateSync.Unlock()
	}
	logInfo("MAIN", "States read in "+time.Since(timer).String())
}
