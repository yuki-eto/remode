package remodel

import (
	"github.com/yuki-eto/remodel/assert"
	"strings"
	"testing"
)

func TestTable(t *testing.T) {
	t.Run("parse_normal", func(t *testing.T) {
		ddl := `
CREATE TABLE IF NOT EXISTS users (
  id BIGINT(20) UNSIGNED NOT NULL AUTO_INCREMENT,
  uuid VARCHAR(40) NOT NULL,
  is_debug_user TINYINT(1) NOT NULL DEFAULT '0',
  created_at DATETIME,
  updated_at DATETIME,
  PRIMARY KEY (id),
  UNIQUE KEY ''unique_key'' (uuid),
  KEY ''key'' (is_debug_user)
);
`
		ddl = strings.Replace(ddl, "''", "`", 4)
		table := &Table{}
		if err := table.Parse(ddl); err != nil {
			t.Fatal(err)
		}

		assert.Equals(t, table.Name, "users")
		assert.False(t, table.IsReadOnly)
		assert.Len(t, table.Columns, 5)
		assert.Len(t, table.Indexes, 3)

		col := table.Columns[0]
		assert.Equals(t, col.Name, "id")
		assert.True(t, col.IsPrimaryKey)
		assert.True(t, col.IsNotNull)
		assert.Equals(t, col.EntityType, Uint64)

		col = table.Columns[2]
		assert.Equals(t, col.Name, "is_debug_user")
		assert.Equals(t, col.EntityType, Bool)

		index := table.Indexes[0]
		assert.Equals(t, index.Name, "PRIMARY")
		assert.True(t, index.IsPrimaryKey)

		index = table.Indexes[1]
		assert.Equals(t, index.Name, "unique_key")
		assert.False(t, index.IsPrimaryKey)
		assert.True(t, index.IsUnique)

		index = table.Indexes[2]
		assert.Equals(t, index.Name, "key")
		assert.False(t, index.IsUnique)
	})

	t.Run("parse_read_only", func(t *testing.T) {
		ddl := `
CREATE TABLE IF NOT EXISTS items (
  id BIGINT(20) UNSIGNED NOT NULL AUTO_INCREMENT,
  name VARCHAR(40) NOT NULL,
  open_at DATETIME,
  PRIMARY KEY (id)
);
`
		table := &Table{}
		if err := table.Parse(ddl); err != nil {
			t.Fatal(err)
		}

		assert.True(t, table.IsReadOnly)
		assert.Equals(t, table.Name, "items")
		assert.Len(t, table.Columns, 3)

		col := table.Columns[2]
		assert.Equals(t, col.Name, "open_at")
		assert.Equals(t, col.EntityType, TimePtr)
	})
}
