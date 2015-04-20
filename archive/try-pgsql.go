package main

import (
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
	"log"
	"time"
)

func main() {

	db, err := sql.Open("postgres", fmt.Sprintf("host=%s user=%s dbname='%s' password=%s port=%s sslmode=disable",
		"localhost", "postgres", "cdr-pusher", "password", "5433"))
	if err != nil {
		panic(err)
	}

	insertTest, err := db.Prepare("INSERT INTO tabletest(field1, field2) VALUES($1, $2)")
	// insertTest, err := db.Prepare("UPDATE tabletest SET field1=? WHERE id=?")
	if err != nil {
		panic(err)
	}
	defer insertTest.Close() // in reality, you should check this call for error

	// sql_drop := `DROP TABLE tabletest`
	// if _, err := db.Exec(sql_drop); err != nil {
	// 	panic(err)
	// }

	sqlCreate := `CREATE TABLE IF NOT EXISTS tabletest
        (id SERIAL, field1 text, field2 text, created timestamp DEFAULT current_timestamp)`
	if _, err := db.Exec(sqlCreate); err != nil {
		panic(err)
	}

	tx, err := db.Begin()
	if err != nil {
		panic(err)
	}

	// Lets insert values
	var res sql.Result
	res, err = tx.Exec("INSERT INTO tabletest (field1, field2) values ('yo', 'uo')")
	if err != nil {
		println("Exec err:", err.Error())
	} else {
		id, err := res.LastInsertId()
		if err != nil {
			println("LastInsertId:", id)
		} else {
			println("Error:", err.Error())
		}
		num, err := res.RowsAffected()
		println("RowsAffected:", num)
	}

	// Batch Insert
	data := []map[string]string{
		{"field1": "1", "field2": "You"},
		{"field1": "2", "field2": "We"},
		{"field1": "3", "field2": "Them"},
	}

	for _, v := range data {
		println("row:", v["field1"], v["field2"])
		res, err = tx.Stmt(insertTest).Exec(v["field1"], v["field2"])
		if err != nil {
			panic(err)
		}
	}

	// # Select from table
	rows, err := tx.Query("SELECT field1, created FROM tabletest")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()
	for rows.Next() {
		var field1 string
		var created time.Time
		if err := rows.Scan(&field1, &created); err != nil {
			log.Fatal(err)
		}
		fmt.Printf("%s - %s\n", field1, created)
	}
	if err = rows.Err(); err != nil {
		log.Fatal(err)
	}

	if err = tx.Commit(); err != nil {
		log.Fatal(err)
	}
}
