package db

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/Heanthor/quill-secure/node/sensor"
	_ "github.com/mattn/go-sqlite3"
	"github.com/rs/zerolog/log"
	"time"
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
	create index if not exists idx_readings_timestamp on readings(ts);
	`); err != nil {
		return nil, err
	}

	return &DB{db: db}, nil
}

func (d *DB) Close() {
	d.db.Close()
}

func (d *DB) RecordAtmosphericMeasurement(mes sensor.AtmosphericDataLine) error {
	log.Debug().Interface("data", mes).Msg("db: RecordAtmosphericMeasurement")
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

// GetRecentStats retrieves stats in reverse chronological order.
func (d *DB) GetRecentStats(days int) ([]sensor.AtmosphericDataLine, error) {
	log.Debug().Msg("db: GetRecentStats")
	var a []sensor.AtmosphericDataLine

	from := time.Now().Add(-time.Hour * time.Duration(24*days)).Unix()
	rows, err := d.db.Query(`
	select
		 ts,
		 temperature,
		 humidity,
		 pressure,
		 altitude,
		 voc_index
	 from readings
	 where ts >= ?
	 `, from)
	if err != nil {
		return nil, fmt.Errorf("GetRecentStats: failed to get rows: %w", err)
	}
	for rows.Next() {
		var (
			adl   sensor.AtmosphericDataLine
			tsInt int64
		)

		if err := rows.Scan(
			&tsInt,
			&adl.Temperature,
			&adl.Humidity,
			&adl.Pressure,
			&adl.Altitude,
			&adl.VOCIndex,
		); err != nil {
			return nil, fmt.Errorf("GetRecentStats: failed to scan: %w", err)
		}

		adl.Timestamp = time.Unix(tsInt, 0)

		a = append(a, adl)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("GetRecentStats: error in iteration: %w", err)
	}

	return a, nil
}
