package dao

import (
	"example/entity"
	"time"

	"github.com/juju/errors"
	"go.knocknote.io/rapidash"
)

type UserByte interface {
	Save(e *entity.UserByte) error
	Delete(e *entity.UserByte) error
	Find() (*entity.UserByte, error)
	FindByID(k0 uint64) (*entity.UserByte, error)
	FindByIDs(k0 []uint64) (entity.UserBytes, error)
}

type UserByteImpl struct {
	tableName    string
	txGetter     func() (*rapidash.Tx, error)
	qb           func() *rapidash.QueryBuilder
	userIDGetter func() uint64
	uqb          func() *rapidash.QueryBuilder
}

func NewUserByte(txGetter func(string) (*rapidash.Tx, error), userIDGetter func() uint64) UserByte {
	return &UserByteImpl{
		qb: func() *rapidash.QueryBuilder {
			return rapidash.NewQueryBuilder("user_bytes")
		},
		tableName: "user_bytes",
		txGetter: func() (*rapidash.Tx, error) {
			return txGetter("user_bytes")
		},
		uqb: func() *rapidash.QueryBuilder {
			return rapidash.NewQueryBuilder("user_bytes").Eq("user_id", userIDGetter())
		},
		userIDGetter: userIDGetter,
	}
}

func (d *UserByteImpl) Save(e *entity.UserByte) error {
	tx, err := d.txGetter()
	if err != nil {
		return errors.Trace(err)
	}
	now := time.Now()
	e.UpdatedAt = &now
	if e.ID == 0 {
		e.UserID = d.userIDGetter()
		e.CreatedAt = &now
		id, err := tx.CreateByTable(d.tableName, e)
		if err != nil {
			return errors.Trace(err)
		}
		e.ID = uint64(id)
		return nil
	}
	b := d.qb().Eq("id", e.ID)
	m := map[string]interface{}{
		"bytes":      e.Bytes,
		"tags":       e.Tags,
		"updated_at": e.UpdatedAt,
	}
	if err := tx.UpdateByQueryBuilder(b, m); err != nil {
		return errors.Trace(err)
	}
	return nil
}

func (d *UserByteImpl) Delete(e *entity.UserByte) error {
	tx, err := d.txGetter()
	if err != nil {
		return errors.Trace(err)
	}
	if e.ID == 0 {
		return errors.New("cannot delete without identifier")
	}
	b := d.qb().Eq("id", e.ID)
	if err := tx.DeleteByQueryBuilder(b); err != nil {
		return errors.Trace(err)
	}
	return nil
}

func (d *UserByteImpl) Find() (*entity.UserByte, error) {
	tx, err := d.txGetter()
	if err != nil {
		return nil, errors.Trace(err)
	}
	b := d.uqb()
	e := &entity.UserByte{}
	if err := tx.FindByQueryBuilder(b, e); err != nil {
		return nil, errors.Trace(err)
	}
	if e.ID == 0 {
		return nil, nil
	}
	return e, nil
}

func (d *UserByteImpl) FindByID(k0 uint64) (*entity.UserByte, error) {
	tx, err := d.txGetter()
	if err != nil {
		return nil, errors.Trace(err)
	}
	b := d.qb().Eq("id", k0)
	e := &entity.UserByte{}
	if err := tx.FindByQueryBuilder(b, e); err != nil {
		return nil, errors.Trace(err)
	}
	if e.ID == 0 {
		return nil, nil
	}
	return e, nil
}

func (d *UserByteImpl) FindByIDs(k0 []uint64) (entity.UserBytes, error) {
	tx, err := d.txGetter()
	if err != nil {
		return nil, errors.Trace(err)
	}
	b := d.qb().In("id", k0)
	e := &entity.UserBytes{}
	if err := tx.FindByQueryBuilder(b, e); err != nil {
		return nil, errors.Trace(err)
	}
	return *e, nil
}
