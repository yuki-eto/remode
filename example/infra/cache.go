package infra

import (
	"database/sql"

	"github.com/juju/errors"
	"go.knocknote.io/rapidash"
)

var r *rapidash.Rapidash

func InitializeCache(servers []string) error {
	cache, err := rapidash.New(
		rapidash.ServerAddrs(servers),
	)
	if err != nil {
		return errors.Trace(err)
	}
	r = cache
	return nil
}

func WarmUpCache(conn *sql.DB, s *rapidash.Struct, isReadOnly bool) error {
	return errors.Trace(r.WarmUp(conn, s, isReadOnly))
}

func CacheTx(conns ...rapidash.Connection) (*rapidash.Tx, error) {
	tx, err := r.Begin(conns...)
	return tx, errors.Trace(err)
}

func FlushCache() error {
	return errors.Trace(r.Flush())
}
