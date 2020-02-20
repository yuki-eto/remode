package infra

import (
	"database/sql"
	"fmt"
	"sync"

	_ "github.com/go-sql-driver/mysql"
	"github.com/juju/errors"
)

var dbConnections = &sync.Map{}

func GetConnection(conf *DBConnectionConfig) (*DBConnection, error) {
	key := conf.ConnectionString()
	loadedConn, exists := dbConnections.Load(key)
	if exists {
		return loadedConn.(*DBConnection), nil
	}

	conn, err := openDBConnection(conf)
	if err != nil {
		return nil, errors.Trace(err)
	}
	dbConnections.Store(key, conn)
	return conn, errors.Trace(err)
}

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
	conf *DBConnectionConfig
}

func openDBConnection(conf *DBConnectionConfig) (*DBConnection, error) {
	conn, err := sql.Open(conf.Driver, conf.ConnectionString())
	if err != nil {
		return nil, errors.Trace(err)
	}
	return &DBConnection{
		DB: conn,
		conf: conf,
	}, nil
}
