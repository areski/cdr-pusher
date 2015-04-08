package main

import (
	log "github.com/Sirupsen/logrus"
	"github.com/astaxie/beego/orm"
	"github.com/manveru/faker"
	_ "github.com/mattn/go-sqlite3"
	"github.com/nu7hatch/gouuid"
	"math/rand"
	"strconv"
	"sync"
	"time"
)

var once sync.Once

// CDR structure used by Beego ORM
type CDR struct {
	Rowid             int64     `orm:"pk;auto;column(rowid)"`
	CallerIDName      string    `orm:"column(caller_id_name)"`
	CallerIDNumber    string    `orm:"column(caller_id_number)"`
	Duration          int       `orm:"column(duration)"`
	StartStamp        time.Time `orm:"auto_now;column(start_stamp)"`
	DestinationNumber string    `orm:"column(destination_number)"`
	Context           string    `orm:"column(context)"`
	AnswerStamp       time.Time `orm:"auto_now;column(answer_stamp)"`
	EndStamp          time.Time `orm:"auto_now;column(end_stamp)"`
	Billsec           int       `orm:"column(billsec)"`
	HangupCause       string    `orm:"column(hangup_cause)"`
	UUID              string    `orm:"column(uuid)"`
	BlegUUID          string    `orm:"column(bleg_uuid)"`
	AccountCode       string    `orm:"column(account_code)"`
}

func (c *CDR) TableName() string {
	return "cdr"
}

// func connectSqliteDB(sqliteDBpath string) {
// 	orm.RegisterDriver("sqlite3", orm.DR_Sqlite)
// 	orm.RegisterDataBase("default", "sqlite3", sqliteDBpath)
// 	orm.RegisterModel(new(CDR))
// }

func random(min, max int) int {
	rand.Seed(time.Now().Unix())
	return rand.Intn(max-min) + min
}

// GenerateCDR creates a certain amount of CDRs to a given SQLite database
func GenerateCDR(sqliteDBpath string, amount int) error {
	once.Do(func() {
		orm.RegisterDriver("sqlite3", orm.DR_Sqlite)
		// TODO: use f.DBFile instead of hardcoded ./sqlitedb/cdr.db
		// orm.RegisterDataBase("default", "sqlite3", "./sqlitedb/cdr.db")
		log.Info("=================>" + sqliteDBpath)
		orm.RegisterDataBase("default", "sqlite3", sqliteDBpath)
		orm.RegisterModel(new(CDR))
	})
	log.Debug("!!! We will populate " + sqliteDBpath + " with " + strconv.Itoa(amount) + " CDRs !!!")
	fake, _ := faker.New("en")

	// connectSqliteDB(sqliteDBpath)
	o := orm.NewOrm()
	// orm.Debug = true
	o.Using("default")

	uuid4, _ := uuid.NewV4()
	cidname := fake.Name()
	cidnum := fake.PhoneNumber()
	dstnum := fake.CellPhoneNumber()
	duration := random(30, 300)
	billsec := duration - 10
	var listcdr = []CDR{}

	for i := 0; i < amount; i++ {
		cdr := CDR{CallerIDName: cidname, CallerIDNumber: cidnum,
			DestinationNumber: dstnum, UUID: uuid4.String(),
			Duration: duration, Billsec: billsec,
			StartStamp: time.Now(), AnswerStamp: time.Now(), EndStamp: time.Now()}
		listcdr = append(listcdr, cdr)
	}

	successNums, err := o.InsertMulti(50, listcdr)
	if err != nil {
		log.Error(err.Error())
		return err
	}
	log.Info("Generate Fake CDRs, inserted: ", successNums)
	return nil
}
