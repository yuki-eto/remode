package remodel

import (
	"errors"
	"strconv"
	"strings"

	"github.com/gertd/go-pluralize"
	"github.com/xwb1989/sqlparser"
)

type ColumnType string

type Table struct {
	Name       string    `yaml:"name"`
	Columns    []*Column `yaml:"columns"`
	Indexes    []*Index  `yaml:"indexes"`
	IsReadOnly bool      `yaml:"is_read_only"`
}

type Column struct {
	Name            string     `yaml:"name"`
	ColumnType      ColumnType `yaml:"column_type"`
	EntityType      EntityType `yaml:"entity_type"`
	Size            uint64     `yaml:"size"`
	IsAutoIncrement bool       `yaml:"is_auto_increment"`
	IsUnsigned      bool       `yaml:"is_unsigned"`
	IsNotNull       bool       `yaml:"is_not_null"`
	DefaultValue    string     `yaml:"default_value"`
	IsPrimaryKey    bool       `yaml:"is_primary_key"`
	UniqueIndexKeys []string   `yaml:"unique_index_keys"`
	IndexKeys       []string   `yaml:"index_keys"`
}

type Index struct {
	Name         string   `yaml:"name"`
	IsPrimaryKey bool     `yaml:"is_primary_key"`
	IsUnique     bool     `yaml:"is_unique"`
	Columns      []string `yaml:"columns"`
}

func (t *Table) Parse(s string) error {
	stmt, err := sqlparser.Parse(s)
	if err != nil {
		return err
	}
	ddl, ok := stmt.(*sqlparser.DDL)
	if !ok {
		return errors.New("not ddl")
	}
	if ddl.Action != "create" {
		return errors.New("not create table")
	}

	p := pluralize.NewClient()
	t.Name = ddl.NewName.Name.String()
	if p.IsSingular(t.Name) {
		return errors.New("not plural table name")
	}

	var primaryIndex *Index
	indexesColumnMap := map[string][]*Index{}
	for _, i := range ddl.TableSpec.Indexes {
		info := i.Info
		index := &Index{
			Name:         info.Name.String(),
			IsPrimaryKey: info.Primary,
			IsUnique:     info.Unique,
			Columns:      []string{},
		}
		for _, c := range i.Columns {
			columnName := c.Column.String()
			if p.IsPlural(columnName) {
				return errors.New("not singular column name")
			}
			index.Columns = append(index.Columns, columnName)
			indexesColumnMap[columnName] = append(indexesColumnMap[columnName], index)
		}
		t.Indexes = append(t.Indexes, index)
		if info.Primary {
			primaryIndex = index
		}
	}
	if primaryIndex == nil {
		return errors.New("need primary key")
	}
	if len(primaryIndex.Columns) > 1 {
		return errors.New("not single primary key")
	}

	for _, c := range ddl.TableSpec.Columns {
		ct := c.Type
		column := &Column{
			Name:            c.Name.String(),
			ColumnType:      ColumnType(ct.Type),
			IsAutoIncrement: bool(ct.Autoincrement),
			IsUnsigned:      bool(ct.Unsigned),
			IsNotNull:       bool(ct.NotNull),
			UniqueIndexKeys: []string{},
			IndexKeys:       []string{},
		}
		if ct.Length != nil {
			size, err := strconv.ParseUint(string(ct.Length.Val), 10, 64)
			if err != nil {
				return err
			}
			column.Size = size
		}
		column.EntityType = column.entityType()
		if ct.Default != nil {
			defaultStr := string(ct.Default.Val)
			if defaultStr != "null" {
				column.DefaultValue = defaultStr
			}
		}
		if indexes, exists := indexesColumnMap[column.Name]; exists {
			for _, i := range indexes {
				if i.IsPrimaryKey {
					column.IsPrimaryKey = true
				} else if i.IsUnique {
					column.UniqueIndexKeys = append(column.UniqueIndexKeys, i.Name)
				} else {
					column.IndexKeys = append(column.IndexKeys, i.Name)
				}
			}
		}
		t.Columns = append(t.Columns, column)
	}

	if !strings.HasPrefix(t.Name, "user_") && t.Name != "users" {
		t.IsReadOnly = true
	}

	return nil
}

func (c *Column) entityType() EntityType {
	switch c.ColumnType {
	case BigInt:
		if c.IsUnsigned {
			return Uint64
		}
		return Int64
	case MediumInt, Int:
		if c.IsUnsigned {
			return Uint32
		}
		return Int32
	case SmallInt:
		if c.IsUnsigned {
			return Uint16
		}
		return Int16
	case TinyInt:
		if c.Size == 1 && (strings.HasPrefix(c.Name, "is_") || strings.HasPrefix(c.Name, "has_")) {
			return Bool
		}
		if c.IsUnsigned {
			return Uint8
		}
		return Int8
	case Float:
		return Float32
	case Double:
		return Float64
	case Bit:
		if c.Size > 32 {
			return Uint64
		} else if c.Size > 16 {
			return Uint32
		} else if c.Size > 8 {
			return Uint16
		}
		return Uint8
	case Date, Datetime, Timestamp:
		return TimePtr
	case Varchar, Text, Enum:
		return String
	case Blob:
		return ByteSlice
	case Set:
		return StringSlice
	}
	return String
}
