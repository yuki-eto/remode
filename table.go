package remodel

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/gertd/go-pluralize"
	"github.com/juju/errors"
	"github.com/xwb1989/sqlparser"
	"gopkg.in/yaml.v2"
)

type ColumnType string

type Table struct {
	Name       string    `yaml:"name"`
	Columns    []*Column `yaml:"columns"`
	Indexes    []*Index  `yaml:"indexes"`
	IsReadOnly bool      `yaml:"is_read_only"`
}

type Tables []*Table

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

func (s *Tables) Output(rootDir string) error {
	schemaDir := filepath.Join(rootDir, "schema")
	sqlDir := filepath.Join(schemaDir, "sql")
	if _, err := os.Stat(sqlDir); os.IsNotExist(err) {
		return errors.Trace(err)
	}
	yamlDir := filepath.Join(schemaDir, "yaml")
	if _, err := os.Stat(yamlDir); os.IsNotExist(err) {
		if err := os.Mkdir(yamlDir, 0755); err != nil {
			return errors.Trace(err)
		}
		log.Printf("create directory: %s", yamlDir)
	}

	if err := filepath.Walk(sqlDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return errors.Trace(err)
		}
		if info.IsDir() {
			return nil
		}
		if filepath.Ext(path) != ".sql" {
			return nil
		}
		b, err := ioutil.ReadFile(path)
		if err != nil {
			return errors.Trace(err)
		}
		t := &Table{}
		if err := t.parse(string(b)); err != nil {
			return errors.Trace(err)
		}
		*s = append(*s, t)
		return nil
	}); err != nil {
		return errors.Trace(err)
	}

	for _, t := range *s {
		ymlPath := filepath.Join(yamlDir, fmt.Sprintf("%s.yml", t.Name))
		f, err := os.Create(ymlPath)
		if err != nil {
			return errors.Trace(err)
		}
		enc := yaml.NewEncoder(f)
		if err := enc.Encode(t); err != nil {
			return errors.Trace(err)
		}
		if err := enc.Close(); err != nil {
			return errors.Trace(err)
		}
		log.Printf("output: %s", ymlPath)
	}

	return nil
}

func (s *Tables) Load(rootPath string) error {
	matches, err := filepath.Glob(filepath.Join(rootPath, "schema", "yaml", "*.yml"))
	if err != nil {
		return errors.Trace(err)
	}

	for _, path := range matches {
		f, err := os.Open(path)
		if err != nil {
			return errors.Trace(err)
		}
		dec := yaml.NewDecoder(f)
		var t *Table
		if err := dec.Decode(&t); err != nil {
			return errors.Trace(err)
		}
		if err := f.Close(); err != nil {
			return errors.Trace(err)
		}
		*s = append(*s, t)
	}

	return nil
}

func (s *Tables) Entities() *Entities {
	es := Entities{}
	for _, t := range *s {
		e := &Entity{}
		e.fromTable(t)
		es = append(es, e)
	}
	return &es
}

func (s *Tables) Daos() *Daos {
	ds := Daos{}
	for _, t := range *s {
		d := &Dao{}
		d.fromTable(t)
		ds = append(ds, d)
	}
	return &ds
}

func (s *Tables) Models() *Models {
	ms := Models{}
	for _, t := range *s {
		m := &Model{}
		m.fromTable(t)
		ms = append(ms, m)
	}
	return &ms
}

func (t *Table) parse(s string) error {
	stmt, err := sqlparser.Parse(s)
	if err != nil {
		return errors.Trace(err)
	}
	ddl, ok := stmt.(*sqlparser.DDL)
	if !ok {
		return errors.New("not ddl")
	}
	if ddl.Action != "create" {
		return errors.New("not create table")
	}
	if ddl.TableSpec == nil {
		return errors.New("cannot find table spec")
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
	case Double, Decimal, Numeric:
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
	case Date, Datetime, Timestamp, Time, Year:
		return TimePtr
	case LongBlob, MediumBlob, TinyBlob, Blob, Binary, VarBinary:
		return ByteSlice
	case Set:
		return StringSlice
	}
	return String
}
