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
	cachedWorkplaceStateRecords = map[int]int{}
	workplaceStateRecordSync    sync.Mutex
)

var (
	cachedWorkplaceDowntimeRecords = map[int]uint{}
	workplaceDowntimeRecordSync    sync.Mutex
)

var (
	cachedWorkplaceOrderRecords = map[int]database.OrderRecord{}
	workplaceOrderRecordSync    sync.Mutex
)

var (
	cachedDeviceWorkplaceRecords = map[uint]database.DeviceWorkplaceRecord{}
	deviceWorkplaceRecordSync    sync.Mutex
)

var (
	cachedWorkplacePorts = map[int][]database.WorkplacePort{}
	workplacePortSync    sync.Mutex
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
		readDeviceWorkplaceRecords()
		readActiveStates()
		readLatestWorkplaceStateRecords()
		readLatestWorkplaceDowntimeRecords()
		readLatestWorkplaceOrderRecords()
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

func readLatestWorkplaceOrderRecords() {
	logInfo("MAIN", "Reading workplace order records")
	timer := time.Now()
	db, err := gorm.Open(postgres.Open(config), &gorm.Config{})
	sqlDB, _ := db.DB()
	defer sqlDB.Close()
	if err != nil {
		logError("MAIN", "Problem opening database: "+err.Error())
		activeDevices = nil
		return
	}
	var orderRecords []database.OrderRecord
	db.Where("date_time_end is null").Find(&orderRecords)
	workplaceOrderRecordSync.Lock()
	cachedWorkplaceOrderRecords = make(map[int]database.OrderRecord)
	for _, record := range orderRecords {
		cachedWorkplaceOrderRecords[record.WorkplaceID] = record
	}
	workplaceOrderRecordSync.Unlock()
	logInfo("MAIN", "Workplace order records read in "+time.Since(timer).String())
}

func readLatestWorkplaceDowntimeRecords() {
	logInfo("MAIN", "Reading workplace downtime records")
	timer := time.Now()
	db, err := gorm.Open(postgres.Open(config), &gorm.Config{})
	sqlDB, _ := db.DB()
	defer sqlDB.Close()
	if err != nil {
		logError("MAIN", "Problem opening database: "+err.Error())
		activeDevices = nil
		return
	}
	var downtimeRecords []database.DowntimeRecord
	db.Where("date_time_end is null").Find(&downtimeRecords)
	workplaceDowntimeRecordSync.Lock()
	cachedWorkplaceDowntimeRecords = make(map[int]uint)
	for _, record := range downtimeRecords {
		cachedWorkplaceDowntimeRecords[record.WorkplaceID] = record.ID
	}
	workplaceDowntimeRecordSync.Unlock()
	logInfo("MAIN", "Workplace downtime records read in "+time.Since(timer).String())
}

func readLatestWorkplaceStateRecords() {
	logInfo("MAIN", "Reading workplace state records")
	timer := time.Now()
	db, err := gorm.Open(postgres.Open(config), &gorm.Config{})
	sqlDB, _ := db.DB()
	defer sqlDB.Close()
	if err != nil {
		logError("MAIN", "Problem opening database: "+err.Error())
		activeDevices = nil
		return
	}
	var stateRecords []database.StateRecord
	db.Raw("select * from state_records where id in (select distinct max(id) as id from state_records group by workplace_id)").Find(&stateRecords)
	workplaceStateRecordSync.Lock()
	for _, record := range stateRecords {
		cachedWorkplaceStateRecords[record.WorkplaceID] = record.StateID
	}
	workplaceStateRecordSync.Unlock()
	logInfo("MAIN", "Workplace state records read in "+time.Since(timer).String())
}

func readActiveStates() {
	logInfo("MAIN", "Reading states")
	timer := time.Now()
	db, err := gorm.Open(postgres.Open(config), &gorm.Config{})
	sqlDB, _ := db.DB()
	defer sqlDB.Close()
	if err != nil {
		logError("MAIN", "Problem opening database: "+err.Error())
		activeDevices = nil
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

func readDeviceWorkplaceRecords() {
	logInfo("MAIN", "Reading device_workplace_records")
	timer := time.Now()
	db, err := gorm.Open(postgres.Open(config), &gorm.Config{})
	sqlDB, _ := db.DB()
	defer sqlDB.Close()
	if err != nil {
		logError("MAIN", "Problem opening database: "+err.Error())
		activeDevices = nil
		return
	}
	var records []database.DeviceWorkplaceRecord
	db.Find(&records)
	if len(records) > 0 {
		for _, record := range records {
			deviceWorkplaceRecordSync.Lock()
			cachedDeviceWorkplaceRecords[uint(record.DeviceID)] = record
			deviceWorkplaceRecordSync.Unlock()
			var workplacePorts []database.WorkplacePort
			db.Where("workplace_id = ?", record.WorkplaceID).Find(&workplacePorts)
			cachedWorkplacePorts[record.WorkplaceID] = workplacePorts
		}

	}
	logInfo("MAIN", "Devices_workplace_records read in "+time.Since(timer).String())
}
