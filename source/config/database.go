package config

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"

	_ "github.com/mattn/go-sqlite3"

	"../tools"
)

type DatabaseDriver string

const (
	SQLiteDriver   DatabaseDriver = "SQLite"
	PostgresDriver DatabaseDriver = "Postgres"
)

type DatabaseConfig struct {
	Driver  DatabaseDriver `json:""`
	Options interface{}    `json:""`
}

type SQLiteOptions struct {
	Path string `json:""`
}

type PostgresOptions struct {
	Host     string `json:""`
	User     string `json:""`
	Password Secret `json:""`
}

func (c *DatabaseConfig) UnmarshalJSON(bytes []byte) error {
	var g struct {
		Driver  DatabaseDriver           `json:""`
		Options tools.InterfaceUnmarshal `json:""`
	}
	if err := json.Unmarshal(bytes, &g); err != nil {
		return err
	}
	switch g.Driver {
	case SQLiteDriver:
		var options SQLiteOptions
		if err := json.Unmarshal(g.Options, &options); err != nil {
			return err
		}
		c.Options = options
	case PostgresDriver:
		var options PostgresOptions
		if err := json.Unmarshal(g.Options, &options); err != nil {
			return err
		}
		c.Options = options
	default:
		return fmt.Errorf("driver '%s' is not supported", g.Driver)
	}
	c.Driver = g.Driver
	return nil
}

func createSQLiteDB(opts SQLiteOptions) (*sql.DB, error) {
	return sql.Open("sqlite3", fmt.Sprintf("file:%s", opts.Path))
}

func createPostgresDB(opts PostgresOptions) (*sql.DB, error) {
	return nil, tools.NotImplementedError
}

func (c *DatabaseConfig) CreateDB() (*sql.DB, error) {
	switch t := c.Options.(type) {
	case SQLiteOptions:
		return createSQLiteDB(t)
	case PostgresOptions:
		return createPostgresDB(t)
	default:
		return nil, errors.New("unsupported database config type")
	}
}
