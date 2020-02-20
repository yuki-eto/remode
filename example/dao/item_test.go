package dao

import (
	"fmt"
	"strings"
	"testing"

	"go.knocknote.io/rapidash"

	"github.com/stretchr/testify/assert"
)

func TestItemImpl(t *testing.T) {
	tx, err := getTxForTest(true)
	if err != nil {
		t.Fatal(err)
	}
	fn := func(string) (*rapidash.Tx, error) {
		return tx, nil
	}

	t.Run("find_all", func(t *testing.T) {
		d := NewItem(fn)
		items, err := d.FindsAll()
		if err != nil {
			t.Fatal(err)
		}
		assert.Len(t, items, 6)

		for i, item := range items {
			assert.Equal(t, item.ID, uint64(i+1))
			assert.True(t, strings.HasPrefix(item.Name, fmt.Sprint(i+1)))
			assert.Equal(t, item.MaxCount, uint16(100))
		}
	})

	t.Run("find_by_id", func(t *testing.T) {
		d := NewItem(fn)
		const id = uint64(1)
		item, err := d.FindByID(id)
		if err != nil {
			t.Fatal(err)
		}
		assert.NotNil(t, item)
		assert.Equal(t, item.ID, id)
		assert.Equal(t, item.MaxCount, uint16(100))
	})

	t.Run("find_by_ids", func(t *testing.T) {
		d := NewItem(fn)
		ids := []uint64{1, 3, 5}
		items, err := d.FindByIDs(ids)
		if err != nil {
			t.Fatal(err)
		}
		assert.Len(t, items, 3)
		for _, item := range items {
			assert.NotNil(t, item)
		}
	})

	t.Run("find_by_type", func(t *testing.T) {
		d := NewItem(fn)
		const typ = "consumable"
		items, err := d.FindByType(typ)
		if err != nil {
			t.Fatal(err)
		}
		assert.Len(t, items, 3)
		for _, item := range items {
			assert.NotNil(t, item)
			assert.Equal(t, item.Type, typ)
		}
	})

	t.Run("find_by_rarity", func(t *testing.T) {
		d := NewItem(fn)
		const rarity = "SR"
		items, err := d.FindByRarity(rarity)
		if err != nil {
			t.Fatal(err)
		}
		assert.Len(t, items, 2)
		for _, item := range items {
			assert.NotNil(t, item)
			assert.Equal(t, item.Rarity, rarity)
		}
	})
}
