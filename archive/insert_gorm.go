package main

import (
	"fmt"
	"github.com/jinzhu/gorm"
	"github.com/manveru/faker"
	_ "github.com/mattn/go-sqlite3"
	"github.com/nu7hatch/gouuid"
	"log"
	"math/rand"
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

func connectSqliteDB() gorm.DB {
	db, err := gorm.Open("sqlite3", "../sqlitedb/cdr.db")
	if err != nil {
		log.Fatal(err)
	}
	return db
}

func fetchCdrSqliteGorm() {
	db := connectSqliteDB()
	// var cdrs []CDRgorm
	var cdrs []map[string]interface{}

	db.Raw("SELECT rowid, caller_id_name, destination_number FROM cdr LIMIT ?", 10).Scan(cdrs)

	// db.Limit(10).Find(&cdrs)
	// fmt.Printf("%s - %v\n", query, cdrs)
	fmt.Println(cdrs)
	fmt.Println("-------------------------------")
}

func random(min, max int) int {
	rand.Seed(time.Now().Unix())
	return rand.Intn(max-min) + min
}

func main() {
	fake, _ := faker.New("en")
	db := connectSqliteDB()
	db.DB().Ping()
	db.LogMode(true)

	var cdrs []CDRgorm
	// var cdrs []map[string]interface{}

	// db.Raw("SELECT rowid, caller_id_name, destination_number FROM cdr LIMIT ?", 10).Scan(cdrs)

	db.Limit(10).Find(&cdrs)
	// fmt.Printf("%s - %v\n", query, cdrs)
	fmt.Println(cdrs)
	fmt.Println("-------------------------------")
	for i := 0; i < 1; i++ {
		uuid4, _ := uuid.NewV4()
		cdr := CDRgorm{CallerIDName: fake.Name(), CallerIDNumber: fake.PhoneNumber(),
			DestinationNumber: fake.CellPhoneNumber(), UUID: uuid4.String(),
			Duration: random(30, 300), Billsec: random(30, 300),
			StartStamp: time.Now(), AnswerStamp: time.Now(), EndStamp: time.Now()}

		fmt.Println(db.NewRecord(cdr))
		db.Create(&cdr)
		fmt.Println(db.NewRecord(cdr))
	}
}
