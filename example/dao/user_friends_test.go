package dao

import (
	"example/entity"
	"example/infra"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.knocknote.io/rapidash"
)

func TestUserFriendsImpl(t *testing.T) {
	conn, err := infra.GetConnection(getDatabaseConfForTest())
	if err != nil {
		t.Fatal(err)
	}
	if _, err := conn.Exec("TRUNCATE TABLE `user_friends`"); err != nil {
		t.Fatal(err)
	}

	getDao := func(t *testing.T, userID uint64) (UserFriend, *rapidash.Tx) {
		tx, err := getTxForTest(false)
		if err != nil {
			t.Fatal(err)
		}
		fn := func(string) (*rapidash.Tx, error) {
			return tx, nil
		}
		idGetter := func() uint64 {
			return userID
		}
		return NewUserFriend(fn, idGetter), tx
	}
	commit := func(t *testing.T, tx *rapidash.Tx) {
		if err := tx.Commit(); err != nil {
			t.Fatal(err)
		}
	}

	t.Run("save", func(t *testing.T) {
		var entities []*entity.UserFriend
		t.Run("insert", func(t *testing.T) {
			d, tx := getDao(t, 0)
			for i := 1; i <= 5; i++ {
				for j := 1; j <= 3; j++ {
					e := &entity.UserFriend{
						UserID:      uint64(i),
						OtherUserID: uint64(j * 10),
					}
					entities = append(entities, e)
					if err := d.Save(e); err != nil {
						t.Fatal(err)
					}
				}
			}
			commit(t, tx)
			for i, e := range entities {
				assert.Equal(t, e.ID, uint64(i+1))
				assert.NotNil(t, e.CreatedAt)
				assert.NotNil(t, e.UpdatedAt)
			}
		})

		t.Run("update", func(t *testing.T) {
			oldUpdateAtMap := map[uint64]*time.Time{}
			d, tx := getDao(t, 0)
			for _, e := range entities {
				e.OtherUserID = e.OtherUserID * 100
				oldUpdateAtMap[e.ID] = e.UpdatedAt
				if err := d.Save(e); err != nil {
					t.Fatal(err)
				}
			}
			commit(t, tx)
			for _, e := range entities {
				assert.True(t, e.UpdatedAt.After(*oldUpdateAtMap[e.ID]))
			}
		})
	})

	t.Run("find", func(t *testing.T) {
		d, tx := getDao(t, 1)
		friends, err := d.Find()
		if err != nil {
			t.Fatal(err)
		}
		commit(t, tx)
		assert.Len(t, friends, 3)
	})
}
