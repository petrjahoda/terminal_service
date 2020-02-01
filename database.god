package main

import (
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"golang.org/x/crypto/bcrypt"
	"log"
	"time"
)

type State struct {
	gorm.Model
	Name  string `gorm:"unique"`
	Color string
	Note  string
}

type DeviceUserRecord struct {
	gorm.Model
	DateTimeStart       time.Time
	DateTimeEnd         time.Time
	Interval            float32
	UserId              uint
	DeviceOrderRecordId uint
	Note                string
}

type DeviceDowntimeRecord struct {
	gorm.Model
	DateTimeStart       time.Time
	DateTimeEnd         time.Time
	Interval            float32
	DowntimeId          uint
	UserId              uint
	DeviceId            uint
	DeviceOrderRecordId uint
	Note                string
}

type DeviceBreakdownRecord struct {
	gorm.Model
	DateTimeStart       time.Time
	DateTimeEnd         time.Time
	Interval            float32
	BreakdownId         uint
	DeviceOrderRecordId uint
	UserId              uint
	DeviceId            uint
	Note                string
}

type DeviceFaultRecord struct {
	gorm.Model
	DateTime            time.Time
	Count               uint
	FaultId             uint
	DeviceOrderRecordId uint
	UserId              uint
	DeviceId            uint
	Note                string
}

type DevicePackageRecord struct {
	gorm.Model
	DateTime            time.Time
	Count               uint
	PackageId           uint
	DeviceId            uint
	UserId              uint
	DeviceOrderRecordId uint
	Note                string
}

type DevicePartRecord struct {
	gorm.Model
	DateTime            time.Time
	Count               uint
	DeviceOrderRecordId uint
	PartId              uint
	UserId              uint
	DeviceId            uint
	Note                string
}

type DeviceOrderRecord struct {
	gorm.Model
	CountOk         uint
	CountNok        uint
	Cavity          uint
	AverageCycle    uint
	DateTimeStart   time.Time
	DateTimeEnd     time.Time
	Interval        float32
	OrderId         uint
	OperationId     uint
	DeviceId        uint
	WorkplaceId     uint
	WorkplaceModeId uint
	WorkshiftId     uint
	Note            string
}

type Operation struct {
	gorm.Model
	Name    string `gorm:"unique"`
	OrderId uint
	Barcode uint
	Note    string
}

type Order struct {
	gorm.Model
	Name            string `gorm:"unique"`
	Barcode         uint
	CountRequest    uint
	Cavity          uint
	ProductId       uint
	WorkplaceId     uint
	DateTimeRequest time.Time
	Note            string
}
type Product struct {
	gorm.Model
	Name             string `gorm:"unique"`
	Barcode          uint
	CycleTime        uint
	DownTimeInterval uint
	Note             string
}

type Part struct {
	gorm.Model
	Name    string `gorm:"unique"`
	Barcode uint
	Note    string
}

type WorkplaceMode struct {
	gorm.Model
	Name             string `gorm:"unique"`
	DowntimeInterval uint
	PoweroffInterval uint
	Note             string
}
type WorkplaceWorkshift struct {
	gorm.Model
	WorkplaceId uint
	WorkshiftId uint
}
type Workshift struct {
	gorm.Model
	Name           string `gorm:"unique"`
	WorkshiftStart time.Time
	WorkshiftEnd   time.Time
	Note           string
}

type User struct {
	gorm.Model
	FirstName  string
	SecondName string
	Login      string
	Barcode    string
	Rfid       string
	Pin        string
	Password   string
	Position   string
	Email      string
	Phone      string
	UserTypeId uint
	UserRoleId uint
	Note       string
}

type UserRole struct {
	gorm.Model
	Name string `gorm:"unique"`
	Note string
}

type UserType struct {
	gorm.Model
	Name string `gorm:"unique"`
	Note string
}

type Downtime struct {
	gorm.Model
	Name           string `gorm:"unique"`
	Barcode        string
	Color          string
	DowntimeTypeId uint
	Note           string
}

type DowntimeType struct {
	gorm.Model
	Name string `gorm:"unique"`
	Note string
}

type Breakdown struct {
	gorm.Model
	Name            string `gorm:"unique"`
	Barcode         string
	Color           string
	BreakdownTypeId uint
	Note            string
}

type BreakdownType struct {
	gorm.Model
	Name string `gorm:"unique"`
	Note string
}

type Fault struct {
	gorm.Model
	Name        string `gorm:"unique"`
	Barcode     string
	FaultTypeId uint
	Note        string
}

type FaultType struct {
	gorm.Model
	Name string `gorm:"unique"`
	Note string
}

type Package struct {
	gorm.Model
	Name          string `gorm:"unique"`
	Barcode       string
	PackageTypeId uint
	OrderId       uint
	Note          string
}

type PackageType struct {
	gorm.Model
	Name  string `gorm:"unique"`
	Count uint
	Note  string
}

type DeviceType struct {
	gorm.Model
	Name string `gorm:"unique"`
	Note string
}

type DevicePortType struct {
	gorm.Model
	Name string `gorm:"unique"`
	Note string
}

type Setting struct {
	gorm.Model
	Key     string `gorm:"unique"`
	Value   string
	Enabled bool
	Note    string
}

type Device struct {
	gorm.Model
	Name         string `gorm:"unique"`
	DeviceTypeId uint
	IpAddress    string `gorm:"unique"`
	MacAddress   string
	TypeName     string
	Activated    bool
	Settings     string
	Workplace    uint
	DevicePorts  []DevicePort
	Note         string
}

type DevicePort struct {
	gorm.Model
	Name               string
	Unit               string
	PortNumber         int
	DevicePortTypeId   uint
	DeviceId           uint
	ActualDataDateTime time.Time
	ActualData         string
	PlcDataType        string
	PlcDataAddress     string
	Settings           string
	Virtual            bool
	Note               string
}

type DeviceAnalogRecord struct {
	Id           uint      `gorm:"primary_key"`
	DevicePortId uint      `gorm:"unique_index:unique_analog_data"`
	DateTime     time.Time `gorm:"unique_index:unique_analog_data"`
	Data         float32
	Interval     float32
}

type DeviceDigitalRecord struct {
	Id           uint      `gorm:"primary_key"`
	DevicePortId uint      `gorm:"unique_index:unique_digital_data"`
	DateTime     time.Time `gorm:"unique_index:unique_digital_data"`
	Data         int
	Interval     float32
}

type DeviceSerialRecord struct {
	Id           uint      `gorm:"primary_key"`
	DevicePortId uint      `gorm:"unique_index:unique_serial_data"`
	DateTime     time.Time `gorm:"unique_index:unique_serial_data"`
	Data         float32
	Interval     float32
}

func CheckDatabase() bool {
	var connectionString string
	var defaultString string
	var dialect string
	if DatabaseType == "postgres" {
		connectionString = "host=" + DatabaseIpAddress + " sslmode=disable port=" + DatabasePort + " user=" + DatabaseLogin + " dbname=" + DatabaseName + " password=" + DatabasePassword
		defaultString = "host=" + DatabaseIpAddress + " sslmode=disable port=" + DatabasePort + " user=" + DatabaseLogin + " dbname=postgres password=" + DatabasePassword
		dialect = "postgres"
	} else if DatabaseType == "mysql" {
		connectionString = DatabaseLogin + ":" + DatabasePassword + "@tcp(" + DatabaseIpAddress + ":" + DatabasePort + ")/" + DatabaseName + "?charset=utf8&parseTime=True&loc=Local"
		defaultString = DatabaseLogin + ":" + DatabasePassword + "@tcp(" + DatabaseIpAddress + ":" + DatabasePort + ")/information_schema?charset=utf8&parseTime=True&loc=Local"
		dialect = "mysql"
	}
	db, err := gorm.Open(dialect, connectionString)
	if err != nil {
		LogWarning("MAIN", "Database zapsi4 does not exist")
		db, err = gorm.Open(dialect, defaultString)
		if err != nil {
			LogError("MAIN", "Problem opening database: "+err.Error())
			return false
		}
		db = db.Exec("CREATE DATABASE zapsi4;")
		if db.Error != nil {
			LogError("MAIN", "Cannot create database zapsi4")
		}
		LogInfo("MAIN", "Database zapsi4 created")
		return true

	}
	defer db.Close()
	LogDebug("MAIN", "Database zapsi4 exists")
	return true
}

func CheckTables() bool {
	connectionString, dialect := CheckDatabaseType()
	db, err := gorm.Open(dialect, connectionString)
	if err != nil {
		LogError("MAIN", "Problem opening "+dialect+" database: "+err.Error())
		return false
	}
	defer db.Close()
	if !db.HasTable(&DeviceType{}) {
		LogInfo("MAIN", "DeviceTypeId table not exists, creating")
		db.CreateTable(&DeviceType{})
		zapsi := DeviceType{Name: "Zapsi"}
		db.NewRecord(zapsi)
		db.Create(&zapsi)
		zapsiTouch := DeviceType{Name: "Zapsi Touch"}
		db.NewRecord(zapsiTouch)
		db.Create(&zapsiTouch)
		siemens := DeviceType{Name: "Siemens"}
		db.NewRecord(siemens)
		db.Create(&siemens)
		opc := DeviceType{Name: "OPC"}
		db.NewRecord(opc)
		db.Create(&opc)
		scale := DeviceType{Name: "Scale"}
		db.NewRecord(scale)
		db.Create(&scale)
		printer := DeviceType{Name: "Printer"}
		db.NewRecord(printer)
		db.Create(&printer)
		fileImport := DeviceType{Name: "File Import"}
		db.NewRecord(fileImport)
		db.Create(&fileImport)
		smtp := DeviceType{Name: "SMTP"}
		db.NewRecord(smtp)
		db.Create(&smtp)
	} else {
		db.AutoMigrate(&DeviceType{})
	}
	if !db.HasTable(&Device{}) {
		LogInfo("MAIN", "Device table not exists, creating")
		db.CreateTable(&Device{})
		db.Model(&Device{}).AddForeignKey("device_type_id", "device_types(id)", "RESTRICT", "RESTRICT")
	} else {
		db.AutoMigrate(&Device{})
	}
	if !db.HasTable(&Setting{}) {
		LogInfo("MAIN", "Setting table not exists, creating")
		db.CreateTable(&Setting{})
		host := Setting{Key: "host", Value: "smtp.forpsi.com"}
		db.NewRecord(host)
		db.Create(&host)
		port := Setting{Key: "port", Value: "587"}
		db.NewRecord(port)
		db.Create(&port)
		username := Setting{Key: "username", Value: "jahoda@zapsi.eu"}
		db.NewRecord(username)
		db.Create(&username)
		password := Setting{Key: "password", Value: "password"}
		db.NewRecord(password)
		db.Create(&password)
		email := Setting{Key: "email", Value: "support@zapsi.eu"}
		db.NewRecord(email)
		db.Create(&email)
	} else {
		db.AutoMigrate(&Setting{})
	}
	if !db.HasTable(&DevicePortType{}) {
		LogInfo("MAIN", "DevicePortType table not exists, creating")
		db.CreateTable(&DevicePortType{})
		digital := DevicePortType{Name: "Digital"}
		db.NewRecord(digital)
		db.Create(&digital)
		analog := DevicePortType{Name: "Analog"}
		db.NewRecord(analog)
		db.Create(&analog)
		serial := DevicePortType{Name: "Serial"}
		db.NewRecord(serial)
		db.Create(&serial)
		special := DevicePortType{Name: "Special"}
		db.NewRecord(special)
		db.Create(&special)
	} else {
		db.AutoMigrate(&DevicePortType{})
	}
	if !db.HasTable(&DevicePort{}) {
		LogInfo("MAIN", "DevicePort table not exists, creating")
		db.CreateTable(&DevicePort{})
		db.Model(&DevicePort{}).AddForeignKey("device_id", "devices(id)", "RESTRICT", "RESTRICT")
		db.Model(&DevicePort{}).AddForeignKey("device_port_type_id", "device_port_types(id)", "RESTRICT", "RESTRICT")
	} else {
		db.AutoMigrate(&DevicePort{})
	}
	if !db.HasTable(&DeviceAnalogRecord{}) {
		LogInfo("MAIN", "DeviceAnalogRecord table not exists, creating")
		db.CreateTable(&DeviceAnalogRecord{})
		db.Model(&DeviceAnalogRecord{}).AddForeignKey("device_port_id", "device_ports(id)", "RESTRICT", "RESTRICT")
	} else {
		db.AutoMigrate(&DeviceAnalogRecord{})
	}
	if !db.HasTable(&DeviceDigitalRecord{}) {
		LogInfo("MAIN", "DeviceDigitalRecord table not exists, creating")
		db.CreateTable(&DeviceDigitalRecord{})
		db.Model(&DeviceDigitalRecord{}).AddForeignKey("device_port_id", "device_ports(id)", "RESTRICT", "RESTRICT")
	} else {
		db.AutoMigrate(&DeviceDigitalRecord{})
	}
	if !db.HasTable(&DeviceSerialRecord{}) {
		LogInfo("MAIN", "DeviceSerialRecord table not exists, creating")
		db.CreateTable(&DeviceSerialRecord{})
		db.Model(&DeviceSerialRecord{}).AddForeignKey("device_port_id", "device_ports(id)", "RESTRICT", "RESTRICT")
	} else {
		db.AutoMigrate(&DeviceSerialRecord{})
	}
	if !db.HasTable(&PackageType{}) {
		LogInfo("MAIN", "PackageType table not exists, creating")
		db.CreateTable(&PackageType{})
	} else {
		db.AutoMigrate(&PackageType{})
	}
	if !db.HasTable(&Package{}) {
		LogInfo("MAIN", "Package table not exists, creating")
		db.CreateTable(&Package{})
		db.Model(&Package{}).AddForeignKey("package_type_id", "package_types(id)", "RESTRICT", "RESTRICT")
	} else {
		db.AutoMigrate(&Package{})
	}

	if !db.HasTable(&FaultType{}) {
		LogInfo("MAIN", "FaultType table not exists, creating")
		db.CreateTable(&FaultType{})
	} else {
		db.AutoMigrate(&FaultType{})
	}
	if !db.HasTable(&Fault{}) {
		LogInfo("MAIN", "Fault table not exists, creating")
		db.CreateTable(&Fault{})
		db.Model(&Fault{}).AddForeignKey("fault_type_id", "fault_types(id)", "RESTRICT", "RESTRICT")
	} else {
		db.AutoMigrate(&Fault{})
	}

	if !db.HasTable(&BreakdownType{}) {
		LogInfo("MAIN", "BreakdownType table not exists, creating")
		db.CreateTable(&BreakdownType{})
	} else {
		db.AutoMigrate(&BreakdownType{})
	}
	if !db.HasTable(&Breakdown{}) {
		LogInfo("MAIN", "Breakdown table not exists, creating")
		db.CreateTable(&Breakdown{})
		db.Model(&Breakdown{}).AddForeignKey("breakdown_type_id", "breakdown_types(id)", "RESTRICT", "RESTRICT")
	} else {
		db.AutoMigrate(&Breakdown{})
	}

	if !db.HasTable(&DowntimeType{}) {
		LogInfo("MAIN", "DowntimeType table not exists, creating")
		db.CreateTable(&DowntimeType{})
		system := DowntimeType{Name: "System"}
		db.NewRecord(system)
		db.Create(&system)
	} else {
		db.AutoMigrate(&DowntimeType{})
	}
	if !db.HasTable(&Downtime{}) {
		LogInfo("MAIN", "Downtime table not exists, creating")
		db.CreateTable(&Downtime{})
		db.Model(&Downtime{}).AddForeignKey("downtime_type_id", "downtime_types(id)", "RESTRICT", "RESTRICT")
		system := DowntimeType{}
		db.Where("Name = ?", "System").Find(&system)
		noReasonDowntime := Downtime{Name: "No reason downtime", DowntimeTypeId: system.ID}
		db.NewRecord(noReasonDowntime)
		db.Create(&noReasonDowntime)
	} else {
		db.AutoMigrate(&Downtime{})
	}

	if !db.HasTable(&UserType{}) {
		LogInfo("MAIN", "UserType table not exists, creating")
		db.CreateTable(&UserType{})
		operator := UserType{Name: "Operator"}
		db.NewRecord(operator)
		db.Create(&operator)
		zapsi := UserType{Name: "Zapsi"}
		db.NewRecord(zapsi)
		db.Create(&zapsi)
	} else {
		db.AutoMigrate(&UserType{})
	}

	if !db.HasTable(&UserRole{}) {
		LogInfo("MAIN", "UserRole table not exists, creating")
		db.CreateTable(&UserRole{})
		admin := UserRole{Name: "Administrator"}
		db.NewRecord(admin)
		db.Create(&admin)
		powerUser := UserRole{Name: "PowerUser"}
		db.NewRecord(powerUser)
		db.Create(&powerUser)
		user := UserRole{Name: "User"}
		db.NewRecord(user)
		db.Create(&user)
	} else {
		db.AutoMigrate(&UserRole{})
	}
	if !db.HasTable(&User{}) {
		LogInfo("MAIN", "User table not exists, creating")
		db.CreateTable(&User{})
		db.Model(&User{}).AddForeignKey("user_type_id", "user_types(id)", "RESTRICT", "RESTRICT")
		db.Model(&User{}).AddForeignKey("user_role_id", "user_roles(id)", "RESTRICT", "RESTRICT")
		userRole := UserRole{}
		db.Where("Name = ?", "Administrator").Find(&userRole)
		userType := UserType{}
		db.Where("Name = ?", "Zapsi").Find(&userType)
		password := hashAndSalt([]byte("54321"))
		zapsiUser := User{FirstName: "Zapsi", SecondName: "Zapsi", Password: password, UserRoleId: userRole.ID, UserTypeId: userType.ID}
		db.NewRecord(zapsiUser)
		db.Create(&zapsiUser)
	} else {
		db.AutoMigrate(&User{})
	}

	if !db.HasTable(&Workshift{}) {
		LogInfo("MAIN", "Workshift table not exists, creating")
		db.CreateTable(&Workshift{})
		firstShiftStart := time.Date(2000, 1, 1, 6, 0, 0, 0, time.Local)
		firstShiftEnd := time.Date(2000, 1, 1, 14, 0, 0, 0, time.Local)
		firstShift := Workshift{Name: "First Shift", WorkshiftStart: firstShiftStart, WorkshiftEnd: firstShiftEnd}
		db.NewRecord(firstShift)
		db.Create(&firstShift)
		secondShiftStart := time.Date(2000, 1, 1, 14, 0, 0, 0, time.Local)
		secondShiftEnd := time.Date(2000, 1, 1, 22, 0, 0, 0, time.Local)
		secondShift := Workshift{Name: "Second Shift", WorkshiftStart: secondShiftStart, WorkshiftEnd: secondShiftEnd}
		db.NewRecord(secondShift)
		db.Create(&secondShift)
		thirdShiftStart := time.Date(2000, 1, 1, 22, 0, 0, 0, time.Local)
		thirdShiftEnd := time.Date(2000, 1, 2, 6, 0, 0, 0, time.Local)
		thirdShift := Workshift{Name: "Third Shift", WorkshiftStart: thirdShiftStart, WorkshiftEnd: thirdShiftEnd}
		db.NewRecord(thirdShift)
		db.Create(&thirdShift)
	} else {
		db.AutoMigrate(&Workshift{})
	}

	if !db.HasTable(&WorkplaceWorkshift{}) {
		LogInfo("MAIN", "WorkplaceWorkshift table not exists, creating")
		db.CreateTable(&WorkplaceWorkshift{})
		db.Model(&WorkplaceWorkshift{}).AddForeignKey("workplace_id", "workplaces(id)", "RESTRICT", "RESTRICT")
		db.Model(&WorkplaceWorkshift{}).AddForeignKey("workshift_id", "workshifts(id)", "RESTRICT", "RESTRICT")
	} else {
		db.AutoMigrate(&Workshift{})
	}

	if !db.HasTable(&WorkplaceMode{}) {
		LogInfo("MAIN", "Workplacemode table not exists, creating")
		db.CreateTable(&WorkplaceMode{})
		mode := WorkplaceMode{Name: "Production", DowntimeInterval: 300, PoweroffInterval: 300}
		db.NewRecord(mode)
		db.Create(&mode)
	} else {
		db.AutoMigrate(&WorkplaceMode{})
	}
	if !db.HasTable(&Part{}) {
		LogInfo("MAIN", "Part table not exists, creating")
		db.CreateTable(&Part{})
	} else {
		db.AutoMigrate(&Part{})
	}

	if !db.HasTable(&Product{}) {
		LogInfo("MAIN", "Product table not exists, creating")
		db.CreateTable(&Product{})
	} else {
		db.AutoMigrate(&Product{})
	}

	if !db.HasTable(&Order{}) {
		LogInfo("MAIN", "Order table not exists, creating")
		db.CreateTable(&Order{})
		db.Model(&Order{}).AddForeignKey("product_id", "products(id)", "RESTRICT", "RESTRICT")
		db.Model(&Order{}).AddForeignKey("workplace_id", "workplaces(id)", "RESTRICT", "RESTRICT")

	} else {
		db.AutoMigrate(&Order{})
	}

	if !db.HasTable(&Operation{}) {
		LogInfo("MAIN", "Operation table not exists, creating")
		db.CreateTable(&Operation{})
		db.Model(&Operation{}).AddForeignKey("order_id", "orders(id)", "RESTRICT", "RESTRICT")
	} else {
		db.AutoMigrate(&Operation{})
	}

	if !db.HasTable(&DevicePartRecord{}) {
		LogInfo("MAIN", "DevicePartRecord table not exists, creating")
		db.CreateTable(&DevicePartRecord{})
		db.Model(&DevicePartRecord{}).AddForeignKey("part_id", "parts(id)", "RESTRICT", "RESTRICT")
		db.Model(&DevicePartRecord{}).AddForeignKey("user_id", "users(id)", "RESTRICT", "RESTRICT")
		db.Model(&DevicePartRecord{}).AddForeignKey("device_id", "devices(id)", "RESTRICT", "RESTRICT")
		db.Model(&DevicePartRecord{}).AddForeignKey("device_order_record_id", "device_order_records(id)", "RESTRICT", "RESTRICT")
	} else {
		db.AutoMigrate(&DevicePartRecord{})
	}

	if !db.HasTable(&DeviceOrderRecord{}) {
		LogInfo("MAIN", "DeviceOrderRecord table not exists, creating")
		db.CreateTable(&DeviceOrderRecord{})
		db.Model(&DeviceOrderRecord{}).AddForeignKey("order_id", "orders(id)", "RESTRICT", "RESTRICT")
		db.Model(&DeviceOrderRecord{}).AddForeignKey("operation_id", "operations(id)", "RESTRICT", "RESTRICT")
		db.Model(&DeviceOrderRecord{}).AddForeignKey("device_id", "devices(id)", "RESTRICT", "RESTRICT")
		db.Model(&DeviceOrderRecord{}).AddForeignKey("workplace_id", "workplaces(id)", "RESTRICT", "RESTRICT")
		db.Model(&DeviceOrderRecord{}).AddForeignKey("workplace_mode_id", "workplace_modes(id)", "RESTRICT", "RESTRICT")
		db.Model(&DeviceOrderRecord{}).AddForeignKey("workshift_id", "workshifts(id)", "RESTRICT", "RESTRICT")
	} else {
		db.AutoMigrate(&DeviceOrderRecord{})
	}

	if !db.HasTable(&DevicePartRecord{}) {
		LogInfo("MAIN", "DevicePackageRecord table not exists, creating")
		db.CreateTable(&DevicePackageRecord{})
		db.Model(&DevicePackageRecord{}).AddForeignKey("package_id", "packages(id)", "RESTRICT", "RESTRICT")
		db.Model(&DevicePackageRecord{}).AddForeignKey("user_id", "users(id)", "RESTRICT", "RESTRICT")
		db.Model(&DevicePackageRecord{}).AddForeignKey("device_id", "devices(id)", "RESTRICT", "RESTRICT")
		db.Model(&DevicePackageRecord{}).AddForeignKey("device_order_id", "device_orders(id)", "RESTRICT", "RESTRICT")
	} else {
		db.AutoMigrate(&DevicePackageRecord{})
	}

	if !db.HasTable(&DeviceFaultRecord{}) {
		LogInfo("MAIN", "DeviceFaultRecord table not exists, creating")
		db.CreateTable(&DeviceFaultRecord{})
		db.Model(&DeviceFaultRecord{}).AddForeignKey("fault_id", "faults(id)", "RESTRICT", "RESTRICT")
		db.Model(&DeviceFaultRecord{}).AddForeignKey("user_id", "users(id)", "RESTRICT", "RESTRICT")
		db.Model(&DeviceFaultRecord{}).AddForeignKey("device_id", "devices(id)", "RESTRICT", "RESTRICT")
		db.Model(&DeviceFaultRecord{}).AddForeignKey("device_order_record_id", "device_order_records(id)", "RESTRICT", "RESTRICT")
	} else {
		db.AutoMigrate(&DeviceFaultRecord{})
	}

	if !db.HasTable(&DeviceBreakdownRecord{}) {
		LogInfo("MAIN", "DeviceBreakdownRecord table not exists, creating")
		db.CreateTable(&DeviceBreakdownRecord{})
		db.Model(&DeviceBreakdownRecord{}).AddForeignKey("breakdown_id", "breakdowns(id)", "RESTRICT", "RESTRICT")
		db.Model(&DeviceBreakdownRecord{}).AddForeignKey("user_id", "users(id)", "RESTRICT", "RESTRICT")
		db.Model(&DeviceBreakdownRecord{}).AddForeignKey("device_id", "devices(id)", "RESTRICT", "RESTRICT")
		db.Model(&DeviceBreakdownRecord{}).AddForeignKey("device_order_record_id", "device_order_records(id)", "RESTRICT", "RESTRICT")
	} else {
		db.AutoMigrate(&DeviceBreakdownRecord{})
	}

	if !db.HasTable(&DeviceDowntimeRecord{}) {
		LogInfo("MAIN", "DeviceDowntimeRecord table not exists, creating")
		db.CreateTable(&DeviceDowntimeRecord{})
		db.Model(&DeviceDowntimeRecord{}).AddForeignKey("downtime_id", "downtimes(id)", "RESTRICT", "RESTRICT")
		db.Model(&DeviceDowntimeRecord{}).AddForeignKey("user_id", "users(id)", "RESTRICT", "RESTRICT")
		db.Model(&DeviceDowntimeRecord{}).AddForeignKey("device_id", "devices(id)", "RESTRICT", "RESTRICT")
		db.Model(&DeviceDowntimeRecord{}).AddForeignKey("device_order_record_id", "device_order_records(id)", "RESTRICT", "RESTRICT")
	} else {
		db.AutoMigrate(&DeviceDowntimeRecord{})
	}

	if !db.HasTable(&DeviceUserRecord{}) {
		LogInfo("MAIN", "DeviceUserRecord table not exists, creating")
		db.CreateTable(&DeviceUserRecord{})
		db.Model(&DeviceUserRecord{}).AddForeignKey("user_id", "users(id)", "RESTRICT", "RESTRICT")
		db.Model(&DeviceUserRecord{}).AddForeignKey("device_order_record_id", "device_order_records(id)", "RESTRICT", "RESTRICT")
	} else {
		db.AutoMigrate(&DeviceUserRecord{})
	}

	return true
}

func CheckDatabaseType() (string, string) {
	var connectionString string
	var dialect string
	if DatabaseType == "postgres" {
		connectionString = "host=" + DatabaseIpAddress + " sslmode=disable port=" + DatabasePort + " user=" + DatabaseLogin + " dbname=" + DatabaseName + " password=" + DatabasePassword
		dialect = "postgres"
	} else if DatabaseType == "mysql" {
		connectionString = DatabaseLogin + ":" + DatabasePassword + "@tcp(" + DatabaseIpAddress + ":" + DatabasePort + ")/" + DatabaseName + "?charset=utf8&parseTime=True&loc=Local"
		dialect = "mysql"
	}
	return connectionString, dialect
}

func hashAndSalt(pwd []byte) string {
	hash, err := bcrypt.GenerateFromPassword(pwd, bcrypt.MinCost)
	if err != nil {
		log.Println(err)
	}
	return string(hash)
}
