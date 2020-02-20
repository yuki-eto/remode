package dao

import (
	"example/entity"
	"example/infra"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.knocknote.io/rapidash"
)

func TestUserImpl(t *testing.T) {
	conn, err := infra.GetConnection(getDatabaseConfForTest())
	if err != nil {
		t.Fatal(err)
	}
	if _, err := conn.Exec("TRUNCATE TABLE `users`"); err != nil {
		t.Fatal(err)
	}

	getDao := func(t *testing.T) (User, *rapidash.Tx) {
		tx, err := getTxForTest(false)
		if err != nil {
			t.Fatal(err)
		}
		fn := func(string) (*rapidash.Tx, error) {
			return tx, nil
		}
		return NewUser(fn), tx
	}
	commit := func(t *testing.T, tx *rapidash.Tx) {
		if err := tx.Commit(); err != nil {
			t.Fatal(err)
		}
	}

	t.Run("save", func(t *testing.T) {
		var entities []*entity.User
		t.Run("insert", func(t *testing.T) {
			d, tx := getDao(t)
			for i := 1; i < 10; i++ {
				e := &entity.User{
					Uuid:          fmt.Sprintf("uuid_%d", i),
					AccessToken:   fmt.Sprintf("token_%d", i),
					OutsideUserID: fmt.Sprintf("ouid_%d", i),
					Name:          fmt.Sprintf("name_%d", i),
				}
				entities = append(entities, e)
				if err := d.Save(e); err != nil {
					t.Fatal(err)
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
			d, tx := getDao(t)
			for _, e := range entities {
				e.Uuid = fmt.Sprintf("uuid_%d", e.ID*10)
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

		t.Run("find_by_id", func(t *testing.T) {
			d, tx := getDao(t)
			const id = uint64(1)
			u, err := d.FindByID(id)
			if err != nil {
				t.Fatal(err)
			}
			commit(t, tx)
			assert.NotNil(t, u)
			assert.Equal(t, u.ID, id)
		})

		t.Run("find_by_ids", func(t *testing.T) {
			d, tx := getDao(t)
			ids := []uint64{1, 3, 5}
			users, err := d.FindByIDs(ids)
			if err != nil {
				t.Fatal(err)
			}
			commit(t, tx)
			assert.Len(t, users, 3)
		})

		t.Run("find_by_uuid", func(t *testing.T) {
			d, tx := getDao(t)
			const uuid = "uuid_10"
			u, err := d.FindByUuid(uuid)
			if err != nil {
				t.Fatal(err)
			}
			commit(t, tx)
			assert.NotNil(t, u)
			assert.Equal(t, uint64(1), u.ID)
			assert.Equal(t, uuid, u.Uuid)
		})
	})

	t.Run("find_by_uuids", func(t *testing.T) {
		d, tx := getDao(t)
		uuids := []string{"uuid_10", "uuid_20", "uuid_30"}
		users, err := d.FindByUuids(uuids)
		if err != nil {
			t.Fatal(err)
		}
		commit(t, tx)
		assert.Len(t, users, 3)
	})

	t.Run("find_by_outside_user_id", func(t *testing.T) {
		d, tx := getDao(t)
		const ouid = "ouid_1"
		u, err := d.FindByOutsideUserID(ouid)
		if err != nil {
			t.Fatal(err)
		}
		commit(t, tx)
		assert.NotNil(t, u)
		assert.Equal(t, uint64(1), u.ID)
		assert.Equal(t, ouid, u.OutsideUserID)
	})

	t.Run("find_by_outside_user_ids", func(t *testing.T) {
		d, tx := getDao(t)
		ouids := []string{"ouid_1", "ouid_2", "ouid_3"}
		users, err := d.FindByOutsideUserIDs(ouids)
		if err != nil {
			t.Fatal(err)
		}
		commit(t, tx)
		assert.Len(t, users, 3)
	})
}
