package dao

import (
	"example/entity"
	"example/infra"
	"log"
	"os"
	"testing"

	"github.com/juju/errors"
	"go.knocknote.io/rapidash"
)

func getDatabaseConfForTest() *infra.DBConnectionConfig {
	return &infra.DBConnectionConfig{
		Driver:   "mysql",
		Username: "root",
		Password: "",
		Host:     "localhost",
		Port:     3306,
		DBName:   "remodel_test",
	}
}

func TestMain(m *testing.M) {
	if err := infra.InitializeCache([]string{"localhost:11211"}); err != nil {
		log.Fatalf("failed to initialize cache: %+v", err)
	}
	if err := infra.FlushCache(); err != nil {
		log.Fatalf("failed to flush cache: %+v", err)
	}

	conn, err := infra.GetConnection(getDatabaseConfForTest())
	if err != nil {
		log.Fatalf("cannot connect to db: %+v", err)
	}

	for _, st := range entity.Structables() {
		if err := infra.WarmUpCache(conn.DB, st.Struct(), false); err != nil {
			log.Fatalf("failed to warm up cache: %+v", err)
		}
	}
	for _, st := range entity.ReadOnlyStructables() {
		if err := infra.WarmUpCache(conn.DB, st.Struct(), true); err != nil {
			log.Fatalf("failed to warm up cache: %+v", err)
		}
	}

	code := m.Run()
	os.Exit(code)
}

func getTxForTest(isOnlyCache bool) (*rapidash.Tx, error) {
	if isOnlyCache {
		rtx, err := infra.CacheTx()
		return rtx, errors.Trace(err)
	}
	conn, err := infra.GetConnection(getDatabaseConfForTest())
	if err != nil {
		return nil, err
	}
	tx, err := conn.Begin()
	if err != nil {
		return nil, err
	}
	rtx, err := infra.CacheTx(tx)
	return rtx, err
}
