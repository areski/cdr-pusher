package fetch

import (
	"fmt"
	"github.com/jinzhu/gorm"
	"github.com/jmoiron/sqlx"
)

// dep:    "github.com/jmoiron/sqlx"
func fetchCdrSqliteSqlx() {
	db, err := sqlx.Open("sqlite3", "./sqlitedb/cdr.db")
	defer db.Close()

	if err != nil {
		fmt.Println("Failed to connect", err)
		return
	}
	fmt.Println("SQLX:> SELECT rowid, caller_id_name, destination_number FROM cdr LIMIT 100")
	// cdrs := make([][]interface{}, 100)
	var cdrs []interface{}
	err = db.Select(&cdrs, "SELECT rowid, caller_id_name, duration FROM cdr LIMIT 100")
	if err != nil {
		fmt.Println("Failed to run query", err)
		return
	}

	fmt.Println(cdrs)
	fmt.Println("-------------------------------")
}

// dep:    "github.com/jinzhu/gorm"
func fetchCdrSqliteGorm() {
	db, err := gorm.Open("sqlite3", "./sqlitedb/cdr.db")
	if err != nil {
		log.Fatal(err)
	}
	// var cdrs []CdrGorm
	var cdrs []map[string]interface{}

	db.Raw("SELECT rowid, caller_id_name, destination_number FROM cdr LIMIT ?", 10).Scan(cdrs)

	// db.Limit(10).Find(&cdrs)
	// fmt.Printf("%s - %v\n", query, cdrs)
	fmt.Println(cdrs)
	fmt.Println("-------------------------------")
}

// CdrGorm used by Gorm to access CDR entity
type CdrGorm struct {
	Rowid          int64     `gorm:"column:rowid"`
	CallerIDName   string    `gorm:"column:caller_id_name"`
	CallerIDNumber string    `gorm:"column:caller_id_number"`
	Duration       int64     `gorm:"column:duration"`
	StartStamp     time.Time `gorm:"column:start_stamp"`
	// destination_number string
	// context            string
	// start_stamp        time.Time
	// answer_stamp       time.Time
	// end_stamp          time.Time
	// duration           int64
	// billsec            int64
	// hangup_cause       string
	// uuid               string
	// bleg_uuid          string
	// account_code       string
}

// TableName define a different table name
func (c CdrGorm) TableName() string {
	return "cdr"
}
