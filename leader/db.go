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
		return nil, errors.New("db cannot use blank filename")
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
