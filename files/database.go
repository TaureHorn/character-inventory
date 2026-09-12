// INITIATE DATABASE CONNECTION AND MANAGE CRUD OPERATIONS
package files

import (
	// "database/sql"
 	_ "github.com/mattn/go-sqlite3"
)

const (
	DEFAULT_SQL string = `
DROP TABLE IF EXISTS album;
CREATE TABLE characters (
    id                  BLOB PRIMARY KEY NOT NULL,
    name                VARCHAR(128) NOT NULL,
    type                VARCHAR(128) NOT NULL,
    inventory           BLOB
);

CREATE TABLE items (
    id                  BLOB PRIMARY KEY NOT NULL,
    sort_value          INTEGER,
    sort_name           VARCHAR(128) NOT NULL,
    name                VARCHAR(128) NOT NULL,
    weight              INTEGER,
    description         TEXT
)`
)
