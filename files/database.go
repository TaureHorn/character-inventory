// INITIATE DATABASE CONNECTION AND MANAGE CRUD OPERATIONS
package files

import (
	"database/sql"
	"reflect"
	"strconv"
	"strings"
	"uuid"

	_ "github.com/mattn/go-sqlite3"
)

const (
	DB_DRIVER      string = "sqlite3"
	NEW_CHARACTER  string = "New Character"
	NEW_ITEM       string = "New Item"
	VALIDATION_TAG string = "validate"
)

const DEFAULT_SQL string = `
DROP TABLE IF EXISTS album;
CREATE TABLE characters (
    uuid                BLOB PRIMARY KEY NOT NULL,
    name                VARCHAR(128) NOT NULL,
    type                VARCHAR(128) NOT NULL,
    inventory           BLOB
);

CREATE TABLE items (
    uuid                BLOB PRIMARY KEY NOT NULL,
    sort_value          INTEGER,
    sort_name           VARCHAR(128) NOT NULL,
    name                VARCHAR(128) NOT NULL,
    weight              INTEGER,
    description         TEXT
)`

func validate(val interface{}) {
	values := reflect.ValueOf(val)
	for index := range values.NumField() {
		field := values.Field(index)
		tag := values.Type().Field(index).Tag.Get(VALIDATION_TAG)
		if tag == "" {
			continue
		}

		// ITERATE OVER TAG RULES
		rules := strings.Split(tag, ",")
		for _, rule := range rules {
			ruleArr := strings.Split(rule, "=")
			switch ruleArr[0] {
			case "init":
				if field.String() == "" {
					// TODO: set field to NEW_CHARACTER
				}
			case "max":
				maxLen, _ := strconv.Atoi(ruleArr[1])
				if len(field.String()) > maxLen {
					// TODO: throw error
				}
			case "min":
				minLen, _ := strconv.Atoi(ruleArr[1])
				if len(field.String()) < minLen {
					// TODO: throw error
				}
			case "required":
				if field.String() == "" {
					// TODO: get field type and make new value of type and set
				}
			}
		}

	}
}

type Character struct {
	Uuid      uuid.UUID `validate:"required"`
	Name      string    `validate:"init=NEW_CHARACTER,min=2,max=32"`
	Type      string
	Inventory []uuid.UUID
}

type Item struct {
	Uuid        uuid.UUID 	`validate:"required"`
	SortValue   uint
	SortName    string
	Name        string 		`validate:"init=NEW_ITEM,min=2,max=32"`
	Weight      uint
	Description string
	Tags        []string
}

type DataHandler struct {
	Context      ValidationContext
	Database     *sql.DB
	Driver       string
	ErrorMessage error
	Filepath     string
}

func (d *DataHandler) Init(driver, filepath string, context ValidationContext) {
	d.Driver = driver
	d.Filepath = filepath
	d.Context = context

	d.openDatabaseConnection()
	return
}

func (d *DataHandler) openDatabaseConnection() {
	d.Database, d.ErrorMessage = sql.Open(d.Driver, d.Filepath)
	return
}

// WARN: Not working
func (d *DataHandler) WriteDefaultData() {
	statement, err := d.Database.Prepare(DEFAULT_SQL)
	if err != nil {
		d.ErrorMessage = err
		return
	}
	statement.Exec()
	return
}
