package infra

import (
	"database/sql"
	"fmt"

	"github.com/juju/errors"
)

type DBConnectionConfig struct {
	Driver   string `yaml:"driver"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	DBName   string `yaml:"db_name"`
}

func (c *DBConnectionConfig) ConnectionString() string {
	return fmt.Sprintf(
		"%s:%s@tcp(%s:%d)/%s?parseTime=True&loc=Asia%%2FTokyo",
		c.Username,
		c.Password,
		c.Host,
		c.Port,
		c.DBName,
	)
}

type DBConnection struct {
	*sql.DB
}

func NewDBConnection(conf *DBConnectionConfig) (*DBConnection, error) {
	conn, err := sql.Open(conf.Driver, conf.ConnectionString())
	if err != nil {
		return nil, errors.Trace(err)
	}
	return &DBConnection{conn}, nil
}
