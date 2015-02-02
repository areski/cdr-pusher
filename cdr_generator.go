package main

import (
	log "github.com/Sirupsen/logrus"
	"github.com/jinzhu/gorm"
	"github.com/manveru/faker"
	_ "github.com/mattn/go-sqlite3"
	"github.com/nu7hatch/gouuid"
	"math/rand"
	"strconv"
	"time"
)

// CDR used by Gorm to access CDR entity
type CDRgorm struct {
	Rowid             int64     `gorm:"column:rowid;primary_key:yes"`
	CallerIDName      string    `gorm:"column:caller_id_name"`
	CallerIDNumber    string    `gorm:"column:caller_id_number"`
	Duration          int       `gorm:"column:duration"`
	StartStamp        time.Time `gorm:"column:start_stamp"`
	DestinationNumber string    `gorm:"column:destination_number"`
	Context           string    `gorm:"column:context"`
	AnswerStamp       time.Time `gorm:"column:answer_stamp"`
	EndStamp          time.Time `gorm:"column:end_stamp"`
	Billsec           int       `gorm:"column:billsec"`
	HangupCause       string    `gorm:"column:hangup_cause"`
	UUID              string    `gorm:"column:uuid"`
	BlegUUID          string    `gorm:"column:bleg_uuid"`
	AccountCode       string    `gorm:"column:account_code"`
}

// TableName define a different table name
func (c CDRgorm) TableName() string {
	return "cdr"
}

func connectSqliteDB(sqliteDBpath string) gorm.DB {
	db, err := gorm.Open("sqlite3", sqliteDBpath)
	if err != nil {
		log.Fatal(err)
	}
	return db
}

func random(min, max int) int {
	rand.Seed(time.Now().Unix())
	return rand.Intn(max-min) + min
}

// GenerateCDR creates a certain amount of CDRs to a given SQLite database
func GenerateCDR(sqliteDBpath string, amount int) error {
	log.Debug("!!! We will populate " + sqliteDBpath + " with " + strconv.Itoa(amount) + " CDRs !!!")
	fake, _ := faker.New("en")
	db := connectSqliteDB(sqliteDBpath)
	db.DB().SetMaxIdleConns(10)
	db.DB().SetMaxOpenConns(100)
	// db.DB().Ping()
	// db.LogMode(true)
	uuid4, _ := uuid.NewV4()
	cidname := fake.Name()
	cidnum := fake.PhoneNumber()
	dstnum := fake.CellPhoneNumber()
	duration := random(30, 300)
	billsec := duration - 10
	var listcdr = []CDRgorm{}

	for i := 0; i < amount; i++ {
		log.Debug(i)
		cdr := CDRgorm{CallerIDName: cidname, CallerIDNumber: cidnum,
			DestinationNumber: dstnum, UUID: uuid4.String(),
			Duration: duration, Billsec: billsec,
			StartStamp: time.Now(), AnswerStamp: time.Now(), EndStamp: time.Now()}
		listcdr = append(listcdr, cdr)
	}
	db.Create(&listcdr)
	return nil
}
