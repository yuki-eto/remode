package remodel

import (
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

		Equals(t, table.Name, "users")
		False(t, table.IsReadOnly)
		Len(t, table.Columns, 5)
		Len(t, table.Indexes, 3)

		col := table.Columns[0]
		Equals(t, col.Name, "id")
		True(t, col.IsPrimaryKey)
		True(t, col.IsNotNull)
		Equals(t, col.EntityType, Uint64)

		col = table.Columns[2]
		Equals(t, col.Name, "is_debug_user")
		Equals(t, col.EntityType, Bool)

		index := table.Indexes[0]
		Equals(t, index.Name, "PRIMARY")
		True(t, index.IsPrimaryKey)

		index = table.Indexes[1]
		Equals(t, index.Name, "unique_key")
		False(t, index.IsPrimaryKey)
		True(t, index.IsUnique)

		index = table.Indexes[2]
		Equals(t, index.Name, "key")
		False(t, index.IsUnique)
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

		True(t, table.IsReadOnly)
		Equals(t, table.Name, "items")
		Len(t, table.Columns, 3)

		col := table.Columns[2]
		Equals(t, col.Name, "open_at")
		Equals(t, col.EntityType, TimePtr)
	})
}
