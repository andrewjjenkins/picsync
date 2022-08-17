package cache

import (
	"database/sql"

	_ "modernc.org/sqlite"
)

const newDatabaseSchema = `
drop table if exists googlephotos;
create table googlephotos (
	Id INTEGER PRIMARY KEY,
	BaseUrl TEXT,
	Sha256 TEXT,
	Md5 TEXT,
	GooglephotosId TEXT,
	LastUpdated INTEGER,
	LastUsed INTEGER
);
drop table if exists nixplay;
create table nixplay (
	Id INTEGER PRIMARY KEY,
	Url TEXT,
	Filename TEXT,
	SortDate TEXT,
	Md5 TEXT,
	NixplayId INTEGER,
	LastUpdated INTEGER,
	LastUsed INTEGER
);
`

func Open() (*sql.DB, error) {
	db, err := sql.Open("sqlite", "picsync-metadata-cache.db")
	if err != nil {
		return nil, err
	}

	rows, err := db.Query("SELECT name FROM sqlite_master WHERE type='table' AND name='googlephotos';")
	if err != nil {
		return nil, err
	}
	var foundIt bool
	for rows.Next() {
		var name string
		rows.Scan(&name)
		if name == "googlephotos" {
			foundIt = true
		}
	}
	if !foundIt {
		// Need initialization
		err := Init(db)
		if err != nil {
			return nil, err
		}
	}
	return db, nil
}

func Init(db *sql.DB) error {
	if _, err := db.Exec(newDatabaseSchema); err != nil {
		return err
	}
	return nil
}
