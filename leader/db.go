package main

import (
	"database/sql"
	"errors"
	"github.com/Heanthor/quill-secure/node/sensor"
	_ "github.com/mattn/go-sqlite3"
	"github.com/rs/zerolog/log"
)

type DB struct {
	db *sql.DB
}

func NewDB(file string) (*DB, error) {
	if file == "" {
		return nil, errors.New("cannot use blank filename")
	}

	db, err := sql.Open("sqlite3", file)
	if err != nil {
		return nil, err
	}

	if _, err = db.Exec(`
	create table if not exists readings(
	    id integer not null primary key,
	    ts integer not null,
	    temperature real,
	    humidity real,
	    pressure real,
	    altitude real,
	    voc_index real 
	);
	`); err != nil {
		return nil, err
	}

	return &DB{db: db}, nil
}

func (d *DB) Close() {
	d.db.Close()
}

func (d *DB) RecordAtmosphericMeasurement(mes sensor.AtmosphericDataLine) error {
	log.Debug().Interface("data", mes).Msg("RecordAtmosphericMeasurement")
	if _, err := d.db.Exec(`
	insert into readings(
	 ts,
	 temperature,
	 humidity,
	 pressure,
	 altitude,
	 voc_index) values (
	?, ?, ?, ?, ?, ?
	)`, mes.Timestamp.Unix(),
		mes.Temperature,
		mes.Humidity,
		mes.Pressure,
		mes.Altitude,
		mes.VOCIndex); err != nil {
		return err
	}

	return nil
}

//
//func dbTest() {
//	os.RemoveAll("./foo.db")
//
//	db, err := sql.Open("sqlite3", "./foo.db")
//	if err != nil {
//		log.Fatal(err)
//	}
//	defer db.Close()
//
//	sqlStmt := `
//	create table foo (id integer not null primary key, name text);
//	delete from foo;
//	`
//	_, err = db.Exec(sqlStmt)
//	if err != nil {
//		log.Printf("%q: %s\n", err, sqlStmt)
//		return
//	}
//
//	tx, err := db.Begin()
//	if err != nil {
//		log.Fatal(err)
//	}
//	stmt, err := tx.Prepare("insert into foo(id, name) values(?, ?)")
//	if err != nil {
//		log.Fatal(err)
//	}
//	defer stmt.Close()
//	for i := 0; i < 100; i++ {
//		_, err = stmt.Exec(i, fmt.Sprintf("こんにちわ世界%03d", i))
//		if err != nil {
//			log.Fatal(err)
//		}
//	}
//	err = tx.Commit()
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	rows, err := db.Query("select id, name from foo")
//	if err != nil {
//		log.Fatal(err)
//	}
//	defer rows.Close()
//	for rows.Next() {
//		var id int
//		var name string
//		err = rows.Scan(&id, &name)
//		if err != nil {
//			log.Fatal(err)
//		}
//		fmt.Println(id, name)
//	}
//	err = rows.Err()
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	stmt, err = db.Prepare("select name from foo where id = ?")
//	if err != nil {
//		log.Fatal(err)
//	}
//	defer stmt.Close()
//	var name string
//	err = stmt.QueryRow("3").Scan(&name)
//	if err != nil {
//		log.Fatal(err)
//	}
//	fmt.Println(name)
//
//	_, err = db.Exec("delete from foo")
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	_, err = db.Exec("insert into foo(id, name) values(1, 'foo'), (2, 'bar'), (3, 'baz')")
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	rows, err = db.Query("select id, name from foo")
//	if err != nil {
//		log.Fatal(err)
//	}
//	defer rows.Close()
//	for rows.Next() {
//		var id int
//		var name string
//		err = rows.Scan(&id, &name)
//		if err != nil {
//			log.Fatal(err)
//		}
//		fmt.Println(id, name)
//	}
//	err = rows.Err()
//	if err != nil {
//		log.Fatal(err)
//	}
//}
