package dao

import (
	"example/entity"
	"time"

	"github.com/juju/errors"
	"go.knocknote.io/rapidash"
)

type UserFriend interface {
	Save(e *entity.UserFriend) error
	Delete(e *entity.UserFriend) error
	Find() (entity.UserFriends, error)
	FindByID(k0 uint64) (*entity.UserFriend, error)
	FindByIDs(k0 []uint64) (entity.UserFriends, error)
	FindByOtherUserID(k0 uint64) (*entity.UserFriend, error)
	FindByOtherUserIDs(k0 []uint64) (entity.UserFriends, error)
}

type UserFriendImpl struct {
	tableName    string
	txGetter     func() (*rapidash.Tx, error)
	qb           func() *rapidash.QueryBuilder
	userIDGetter func() uint64
	uqb          func() *rapidash.QueryBuilder
}

func NewUserFriend(txGetter func(string) (*rapidash.Tx, error), userIDGetter func() uint64) UserFriend {
	return &UserFriendImpl{
		qb: func() *rapidash.QueryBuilder {
			return rapidash.NewQueryBuilder("user_friends")
		},
		tableName: "user_friends",
		txGetter: func() (*rapidash.Tx, error) {
			return txGetter("user_friends")
		},
		uqb: func() *rapidash.QueryBuilder {
			return rapidash.NewQueryBuilder("user_friends").Eq("user_id", userIDGetter())
		},
		userIDGetter: userIDGetter,
	}
}

func (d *UserFriendImpl) Save(e *entity.UserFriend) error {
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
		"other_user_id": e.OtherUserID,
		"updated_at":    e.UpdatedAt,
	}
	if err := tx.UpdateByQueryBuilder(b, m); err != nil {
		return errors.Trace(err)
	}
	return nil
}

func (d *UserFriendImpl) Delete(e *entity.UserFriend) error {
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

func (d *UserFriendImpl) Find() (entity.UserFriends, error) {
	tx, err := d.txGetter()
	if err != nil {
		return nil, errors.Trace(err)
	}
	b := d.uqb()
	e := &entity.UserFriends{}
	if err := tx.FindByQueryBuilder(b, e); err != nil {
		return nil, errors.Trace(err)
	}
	return *e, nil
}

func (d *UserFriendImpl) FindByID(k0 uint64) (*entity.UserFriend, error) {
	tx, err := d.txGetter()
	if err != nil {
		return nil, errors.Trace(err)
	}
	b := d.qb().Eq("id", k0)
	e := &entity.UserFriend{}
	if err := tx.FindByQueryBuilder(b, e); err != nil {
		return nil, errors.Trace(err)
	}
	if e.ID == 0 {
		return nil, nil
	}
	return e, nil
}

func (d *UserFriendImpl) FindByIDs(k0 []uint64) (entity.UserFriends, error) {
	tx, err := d.txGetter()
	if err != nil {
		return nil, errors.Trace(err)
	}
	b := d.qb().In("id", k0)
	e := &entity.UserFriends{}
	if err := tx.FindByQueryBuilder(b, e); err != nil {
		return nil, errors.Trace(err)
	}
	return *e, nil
}

func (d *UserFriendImpl) FindByOtherUserID(k0 uint64) (*entity.UserFriend, error) {
	tx, err := d.txGetter()
	if err != nil {
		return nil, errors.Trace(err)
	}
	b := d.uqb().Eq("other_user_id", k0)
	e := &entity.UserFriend{}
	if err := tx.FindByQueryBuilder(b, e); err != nil {
		return nil, errors.Trace(err)
	}
	if e.ID == 0 {
		return nil, nil
	}
	return e, nil
}

func (d *UserFriendImpl) FindByOtherUserIDs(k0 []uint64) (entity.UserFriends, error) {
	tx, err := d.txGetter()
	if err != nil {
		return nil, errors.Trace(err)
	}
	b := d.qb().In("other_user_id", k0)
	e := &entity.UserFriends{}
	if err := tx.FindByQueryBuilder(b, e); err != nil {
		return nil, errors.Trace(err)
	}
	return *e, nil
}
