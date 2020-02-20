package dao

import (
	"example/entity"

	"github.com/juju/errors"
	"go.knocknote.io/rapidash"
)

type Item interface {
	FindsAll() (entity.Items, error)
	FindByID(k0 uint64) (*entity.Item, error)
	FindByIDs(k0 []uint64) (entity.Items, error)
	FindByType(k0 string) (entity.Items, error)
	FindByTypes(k0 []string) (entity.Items, error)
	FindByRarity(k0 string) (entity.Items, error)
	FindByRarities(k0 []string) (entity.Items, error)
}

type ItemImpl struct {
	tableName string
	txGetter  func() (*rapidash.Tx, error)
	qb        func() *rapidash.QueryBuilder
}

func NewItem(txGetter func(string) (*rapidash.Tx, error)) Item {
	return &ItemImpl{
		qb: func() *rapidash.QueryBuilder {
			return rapidash.NewQueryBuilder("items")
		},
		tableName: "items",
		txGetter: func() (*rapidash.Tx, error) {
			return txGetter("items")
		},
	}
}

func (d *ItemImpl) FindsAll() (entity.Items, error) {
	tx, err := d.txGetter()
	if err != nil {
		return nil, errors.Trace(err)
	}
	e := &entity.Items{}
	if err := tx.FindAllByTable(d.tableName, e); err != nil {
		return nil, errors.Trace(err)
	}
	return *e, nil
}

func (d *ItemImpl) FindByID(k0 uint64) (*entity.Item, error) {
	tx, err := d.txGetter()
	if err != nil {
		return nil, errors.Trace(err)
	}
	b := d.qb().Eq("id", k0)
	e := &entity.Item{}
	if err := tx.FindByQueryBuilder(b, e); err != nil {
		return nil, errors.Trace(err)
	}
	if e.ID == 0 {
		return nil, nil
	}
	return e, nil
}

func (d *ItemImpl) FindByIDs(k0 []uint64) (entity.Items, error) {
	tx, err := d.txGetter()
	if err != nil {
		return nil, errors.Trace(err)
	}
	b := d.qb().In("id", k0)
	e := &entity.Items{}
	if err := tx.FindByQueryBuilder(b, e); err != nil {
		return nil, errors.Trace(err)
	}
	return *e, nil
}

func (d *ItemImpl) FindByType(k0 string) (entity.Items, error) {
	tx, err := d.txGetter()
	if err != nil {
		return nil, errors.Trace(err)
	}
	b := d.qb().Eq("type", k0)
	e := &entity.Items{}
	if err := tx.FindByQueryBuilder(b, e); err != nil {
		return nil, errors.Trace(err)
	}
	return *e, nil
}

func (d *ItemImpl) FindByTypes(k0 []string) (entity.Items, error) {
	tx, err := d.txGetter()
	if err != nil {
		return nil, errors.Trace(err)
	}
	b := d.qb().In("type", k0)
	e := &entity.Items{}
	if err := tx.FindByQueryBuilder(b, e); err != nil {
		return nil, errors.Trace(err)
	}
	return *e, nil
}

func (d *ItemImpl) FindByRarity(k0 string) (entity.Items, error) {
	tx, err := d.txGetter()
	if err != nil {
		return nil, errors.Trace(err)
	}
	b := d.qb().Eq("rarity", k0)
	e := &entity.Items{}
	if err := tx.FindByQueryBuilder(b, e); err != nil {
		return nil, errors.Trace(err)
	}
	return *e, nil
}

func (d *ItemImpl) FindByRarities(k0 []string) (entity.Items, error) {
	tx, err := d.txGetter()
	if err != nil {
		return nil, errors.Trace(err)
	}
	b := d.qb().In("rarity", k0)
	e := &entity.Items{}
	if err := tx.FindByQueryBuilder(b, e); err != nil {
		return nil, errors.Trace(err)
	}
	return *e, nil
}
