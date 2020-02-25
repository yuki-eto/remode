package remodel

import (
	"fmt"
	"io"
	"strings"

	"github.com/gertd/go-pluralize"
	"github.com/iancoleman/strcase"
	"github.com/juju/errors"
)

type Dao struct {
	Name       string
	TableName  string
	SliceName  string
	Indexes    []*DaoIndex
	Fields     []*DaoField
	IsReadOnly bool
}

type DaoIndex struct {
	*Index
	Columns []*Column
}

type DaoFindMethod struct {
	Name        string
	Args        []code
	IsSliceArg  bool
	IsSlice     bool
	ReturnType  code
	FindColumns []string
}

func (d *DaoIndex) FindMethods(moduleName, entityName, sliceName string, p *pluralize.Client) []*DaoFindMethod {
	var methods []*DaoFindMethod

	var (
		columns    []*Column
		fieldNames []string
	)
	indexSize := len(d.Columns)
	entityPackage := fmt.Sprintf("%s/%s", moduleName, EntityPackageName)
	returnTypeSingle := ptr(qual(entityPackage, entityName))
	returnTypeSlice := qual(entityPackage, sliceName)
	for _, c := range d.Columns {
		columns = append(columns, c)
		fieldName := strcase.ToCamel(c.Name)
		if strings.HasSuffix(fieldName, "Id") {
			l := len(fieldName)
			fieldName = fieldName[:l-2] + "ID"
		}
		if fieldName == "UserID" {
			continue
		}
		fieldNames = append(fieldNames, fieldName)

		m := &DaoFindMethod{
			Name:        "FindBy" + strings.Join(fieldNames, "And"),
			Args:        []code{},
			IsSliceArg:  false,
			IsSlice:     false,
			ReturnType:  returnTypeSingle,
			FindColumns: []string{},
		}
		if !d.IsUnique || len(columns) != indexSize {
			m.IsSlice = true
			m.ReturnType = returnTypeSlice
		}
		j := 0
		for _, col := range columns {
			m.FindColumns = append(m.FindColumns, col.Name)
			if col.Name == "user_id" {
				continue
			}
			m.Args = append(m.Args, i(fmt.Sprintf("k%d", j)).Id(string(col.EntityType)))
			j++
		}
		methods = append(methods, m)

		if len(columns) == 1 {
			findInField := p.Plural(fieldName)
			if strings.HasSuffix(fieldName, "ID") {
				findInField = fieldName + "s"
			}
			m := &DaoFindMethod{
				Name:        "FindBy" + findInField,
				Args:        []code{i("k0").Index().Id(string(c.EntityType))},
				IsSliceArg:  true,
				IsSlice:     true,
				ReturnType:  returnTypeSlice,
				FindColumns: []string{c.Name},
			}
			methods = append(methods, m)
		}
	}

	return methods
}

type DaoField struct {
	Name       string
	ColumnName string
}

func (d *Dao) FromTable(t *Table) {
	columnMap := map[string]*Column{}
	for _, c := range t.Columns {
		columnMap[c.Name] = c
		field := &DaoField{
			ColumnName: c.Name,
		}
		fieldName := strcase.ToCamel(c.Name)
		if strings.HasSuffix(fieldName, "Id") {
			l := len(fieldName)
			fieldName = fieldName[:l-2] + "ID"
		}
		field.Name = fieldName
		d.Fields = append(d.Fields, field)
	}

	p := pluralize.NewClient()
	d.Name = strcase.ToCamel(p.Singular(t.Name))
	d.SliceName = strcase.ToCamel(t.Name)
	d.TableName = t.Name
	d.IsReadOnly = t.IsReadOnly

	d.Indexes = []*DaoIndex{}
	for _, i := range t.Indexes {
		daoIndex := &DaoIndex{
			Index:   i,
			Columns: []*Column{},
		}
		for _, cName := range i.Columns {
			daoIndex.Columns = append(daoIndex.Columns, columnMap[cName])
		}
		d.Indexes = append(d.Indexes, daoIndex)
	}
}

func (d *Dao) GenerateCode(writer io.Writer, moduleName string) error {
	p := pluralize.NewClient()
	f := newFile("dao")

	entityPackage := fmt.Sprintf("%s/%s", moduleName, EntityPackageName)
	f.ImportName(entityPackage, EntityPackageName)
	f.ImportName(RapidashLib, "rapidash")
	f.ImportName(LogLib, "log")
	f.ImportName(ErrorsLib, "errors")

	singleEntity := qual(entityPackage, d.Name)
	ptrEntity := ptr(singleEntity)
	addrEntity := addr(singleEntity)
	entityParam := i("e").Add(ptrEntity)
	sliceEntity := qual(entityPackage, d.SliceName)
	addrSlice := addr(sliceEntity)
	sliceAndError := list(sliceEntity, jerr())

	returnNil := rtn().Nil()
	returnErr := rtn(traceErr())
	returnNilAndErr := rtn(null(), traceErr())

	// define interface
	var methodDefines []code
	if d.IsReadOnly {
		methodDefines = []code{
			i("FindsAll").Params().Params(sliceAndError),
		}
	} else {
		methodDefines = []code{
			i("Save").Params(entityParam).Error(),
			i("Delete").Params(entityParam).Error(),
		}
	}

	var findMethods []*DaoFindMethod
	if strings.HasPrefix(d.TableName, "user_") {
		hasSlice := true
		returnType := sliceEntity
		for _, index := range d.Indexes {
			// user_id でユニークになる？
			if index.IsUnique && len(index.Columns) == 1 && index.Columns[0].Name == "user_id" {
				hasSlice = false
				returnType = ptrEntity
				break
			}
		}
		findMethods = append(findMethods, &DaoFindMethod{
			Name:        "Find",
			Args:        []code{},
			IsSliceArg:  false,
			IsSlice:     hasSlice,
			ReturnType:  returnType,
			FindColumns: []string{"user_id"},
		})
	}
	findMethodNames := map[string]struct{}{}
	for _, index := range d.Indexes {
		mds := index.FindMethods(moduleName, d.Name, d.SliceName, p)
		for _, m := range mds {
			if _, exists := findMethodNames[m.Name]; exists {
				continue
			}
			findMethods = append(findMethods, m)
			findMethodNames[m.Name] = struct{}{}
		}
	}

	structName := d.Name + "Impl"
	for _, m := range findMethods {
		methodDefines = append(methodDefines, i(m.Name).Params(m.Args...).Params(m.ReturnType, jerr()))
	}
	f.Type().Id(d.Name).Interface(methodDefines...).Line()

	qb := qual(RapidashLib, "QueryBuilder")
	structFields := []code{
		i("tableName").String(),
		i("txGetter").Func().Call().Params(ptr(qual(RapidashLib, "Tx")), jerr()),
		i("qb").Func().Call().Params(ptr(qb)),
	}
	structMap := cmap{
		i("tableName"): lit(d.TableName),
		i("txGetter"): fn().Call().Params(ptr(qual(RapidashLib, "Tx")), jerr()).Block(
			rtn(i("txGetter").Call(lit(d.TableName))),
		),
		i("qb"): fn().Call().Params(ptr(qb)).Block(
			rtn(qual(RapidashLib, "NewQueryBuilder").Call(lit(d.TableName))),
		),
	}
	isUserTable := !d.IsReadOnly && strings.HasPrefix(d.TableName, "user_")
	if isUserTable {
		structFields = append(
			structFields,
			i("userIDGetter").Func().Call().Params(i("uint64")),
			i("uqb").Func().Call().Params(ptr(qb)),
		)
		structMap[i("userIDGetter")] = i("userIDGetter")
		structMap[i("uqb")] = fn().Call().Params(ptr(qb)).Block(
			rtn(qual(RapidashLib, "NewQueryBuilder").Call(lit(d.TableName)).Dot("Eq").Call(lit("user_id"), i("userIDGetter").Call())),
		)
	}

	// define impl struct
	f.Type().Id(structName).Struct(structFields...).Line()

	// define methods
	// instantiate
	txGetter := fn().Call(str()).Params(ptr(qual(RapidashLib, "Tx")), jerr())
	params := []code{
		i("txGetter").Add(txGetter),
	}
	if isUserTable {
		params = append(params, i("userIDGetter").Func().Call().Params(i("uint64")))
	}
	f.Func().Id("New" + d.Name).Params(params...).Id(d.Name).Block(
		rtn(addr(i(structName)).Add(vals(structMap))),
	).Line()

	tableName := i("d").Dot("tableName")
	instantiateEntity := i("e").Op(":=").Add(addrEntity).Values()
	instantiateSlice := i("e").Op(":=").Add(addrSlice).Values()
	txGetterCall := list(i("tx"), i("err")).Op(":=").Id("d").Dot("txGetter").Call()
	checkErrAndReturnErr := ifErr().Block(returnErr)
	checkErrAndReturnNilAndErr := ifErr().Block(returnNilAndErr)
	queryBuilder := i("b").Op(":=").Id("d").Dot("qb").Call()
	userQueryBuilder := i("b").Op(":=").Id("d").Dot("uqb").Call()
	idQueryBuilder := queryBuilder.Clone().Dot("Eq").Call(lit("id"), i("e").Dot("ID"))
	tx := i("tx")
	if d.IsReadOnly {
		// FindsAll
		f.Add(pfn("d", structName).Id("FindsAll").Params().Params(sliceAndError).Block(
			txGetterCall,
			checkErrAndReturnNilAndErr,
			instantiateSlice,
			ifxErr(tx.Clone().Dot("FindAllByTable").Call(tableName, i("e"))).Block(
				returnNilAndErr,
			),
			rtn(ptr(i("e")), null()),
		)).Line()
	} else {
		// Save
		m := cmap{}
		var userIDSetter code
		if isUserTable {
			userIDSetter = i("e").Dot("UserID").Op("=").Id("d").Dot("userIDGetter").Call()
		}
		for _, field := range d.Fields {
			if field.ColumnName == "id" || field.ColumnName == "created_at" || field.ColumnName == "user_id" {
				continue
			}
			m[lit(field.ColumnName)] = i("e").Dot(field.Name)
		}
		f.Add(pfn("d", structName).Id("Save").Params(entityParam).Error().Block(
			txGetterCall,
			checkErrAndReturnErr,
			i("now").Op(":=").Qual("time", "Now").Call(),
			i("e").Dot("UpdatedAt").Op("=").Add(addr(i("now"))),
			ifa(i("e").Dot("ID"), "==", lit(0)).Block(
				userIDSetter,
				i("e").Dot("CreatedAt").Op("=").Add(addr(i("now"))),
				list(i("id"), i("err")).Op(":=").Add(tx.Clone().Dot("CreateByTable").Call(tableName, i("e"))),
				ifErr().Block(returnErr),
				i("e").Dot("ID").Op("=").Id("uint64").Params(i("id")),
				returnNil,
			),
			idQueryBuilder,
			i("m").Op(":=").Map(str()).Interface().Add(vals(m)),
			ifxErr(tx.Clone().Dot("UpdateByQueryBuilder").Call(i("b"), i("m"))).Block(
				returnErr,
			),
			returnNil,
		)).Line()

		// Delete
		f.Add(pfn("d", structName).Id("Delete").Params(entityParam).Error().Block(
			txGetterCall,
			checkErrAndReturnErr,
			ifa(i("e").Dot("ID"), "==", lit(0)).Block(
				rtn(qual(ErrorsLib, "New").Call(lit("cannot delete without identifier"))),
			),
			idQueryBuilder,
			ifxErr(tx.Clone().Dot("DeleteByQueryBuilder").Call(i("b"))).Block(
				returnErr,
			),
			returnNil,
		)).Line()
	}

	// findMethods
	for _, m := range findMethods {
		codes := []code{
			txGetterCall,
			checkErrAndReturnNilAndErr,
		}
		q := queryBuilder.Clone()
		if m.FindColumns[0] == "user_id" {
			q = userQueryBuilder.Clone()
		}
		if m.IsSliceArg {
			q.Dot("In").Call(lit(m.FindColumns[0]), i("k0"))
		} else {
			j := 0
			for _, c := range m.FindColumns {
				if c == "user_id" {
					continue
				}
				q.Dot("Eq").Call(lit(c), i(fmt.Sprintf("k%d", j)))
				j++
			}
		}
		codes = append(codes, q)
		if m.IsSlice {
			codes = append(codes, instantiateSlice)
		} else {
			codes = append(codes, instantiateEntity)
		}
		codes = append(codes, ifxErr(tx.Clone().Dot("FindByQueryBuilder").Call(i("b"), i("e"))).Block(returnNilAndErr))
		if m.IsSlice {
			codes = append(codes, rtn(ptr(i("e")), null()))
		} else {
			codes = append(codes, ifa(i("e").Dot("ID"), "==", lit(0)).Block(
				rtn(null(), null()),
			))
			codes = append(codes, rtn(i("e"), null()))
		}
		f.Add(pfn("d", structName).Id(m.Name).Params(m.Args...).Params(m.ReturnType, jerr()).Block(codes...)).Line()
	}

	return errors.Trace(f.Render(writer))
}
