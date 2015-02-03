package main

import (
	"fmt"
	"github.com/astaxie/beego/orm"
	"github.com/manveru/faker"
	_ "github.com/mattn/go-sqlite3"
	"github.com/nu7hatch/gouuid"
	// "log"
	"math/rand"
	"time"
)

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

func init() {
	orm.RegisterDriver("sqlite3", orm.DR_Sqlite)
	orm.RegisterDataBase("default", "sqlite3", "../sqlitedb/cdr.db")
	orm.RegisterModel(new(CDR))
}

func random(min, max int) int {
	rand.Seed(time.Now().Unix())
	return rand.Intn(max-min) + min
}

func main() {
	fake, _ := faker.New("en")

	o := orm.NewOrm()
	o.Using("default")

	fmt.Println("-------------------------------")
	var listcdr = []CDR{}
	var cdr CDR
	for i := 0; i < 100; i++ {
		uuid4, _ := uuid.NewV4()
		cdr = CDR{CallerIDName: fake.Name(), CallerIDNumber: fake.PhoneNumber(),
			DestinationNumber: fake.CellPhoneNumber(), UUID: uuid4.String(),
			Duration: random(30, 300), Billsec: random(30, 300),
			StartStamp: time.Now(), AnswerStamp: time.Now(), EndStamp: time.Now()}
		listcdr = append(listcdr, cdr)
	}

	successNums, err := o.InsertMulti(50, listcdr)
	fmt.Printf("ID: %d, ERR: %v\n", successNums, err)

	// fmt.Println("listcdr:\n%# v\n\n", listcdr)
}
