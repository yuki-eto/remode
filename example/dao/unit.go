package dao

import (
	"github.com/juju/errors"
	"github.com/yuki-eto/remodel/example/entity"
	"go.knocknote.io/rapidash"
)

type Unit interface {
	FindsAll() (entity.Units, error)
	FindByID(k0 uint64) (*entity.Unit, error)
	FindByIDs(k0 []uint64) (entity.Units, error)
	FindByType(k0 string) (entity.Units, error)
	FindByTypes(k0 []string) (entity.Units, error)
	FindByRarity(k0 string) (entity.Units, error)
	FindByRarities(k0 []string) (entity.Units, error)
}

type UnitImpl struct {
	tableName string
	tx        *rapidash.Tx
	qb        func() *rapidash.QueryBuilder
}

func NewUnit(tx *rapidash.Tx, userID uint64) Unit {
	return &UnitImpl{
		qb: func() *rapidash.QueryBuilder {
			return rapidash.NewQueryBuilder("units")
		},
		tableName: "units",
		tx:        tx,
	}
}

func (d *UnitImpl) FindsAll() (entity.Units, error) {
	e := &entity.Units{}
	if err := d.tx.FindAllByTable(d.tableName, e); err != nil {
		return nil, errors.Trace(err)
	}
	return *e, nil
}

func (d *UnitImpl) FindByID(k0 uint64) (*entity.Unit, error) {
	b := d.qb().Eq("id", k0)
	e := &entity.Unit{}
	if err := d.tx.FindByQueryBuilder(b, e); err != nil {
		return nil, errors.Trace(err)
	}
	if e.ID == 0 {
		return nil, nil
	}
	return e, nil
}

func (d *UnitImpl) FindByIDs(k0 []uint64) (entity.Units, error) {
	b := d.qb().In("id", k0)
	e := &entity.Units{}
	if err := d.tx.FindByQueryBuilder(b, e); err != nil {
		return nil, errors.Trace(err)
	}
	return *e, nil
}

func (d *UnitImpl) FindByType(k0 string) (entity.Units, error) {
	b := d.qb().Eq("type", k0)
	e := &entity.Units{}
	if err := d.tx.FindByQueryBuilder(b, e); err != nil {
		return nil, errors.Trace(err)
	}
	return *e, nil
}

func (d *UnitImpl) FindByTypes(k0 []string) (entity.Units, error) {
	b := d.qb().In("type", k0)
	e := &entity.Units{}
	if err := d.tx.FindByQueryBuilder(b, e); err != nil {
		return nil, errors.Trace(err)
	}
	return *e, nil
}

func (d *UnitImpl) FindByRarity(k0 string) (entity.Units, error) {
	b := d.qb().Eq("rarity", k0)
	e := &entity.Units{}
	if err := d.tx.FindByQueryBuilder(b, e); err != nil {
		return nil, errors.Trace(err)
	}
	return *e, nil
}

func (d *UnitImpl) FindByRarities(k0 []string) (entity.Units, error) {
	b := d.qb().In("rarity", k0)
	e := &entity.Units{}
	if err := d.tx.FindByQueryBuilder(b, e); err != nil {
		return nil, errors.Trace(err)
	}
	return *e, nil
}
