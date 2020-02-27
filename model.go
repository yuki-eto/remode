package remodel

import (
	"fmt"
	"io"
	"strings"

	"github.com/gertd/go-pluralize"
	"github.com/iancoleman/strcase"
	"github.com/juju/errors"
)

type Model struct {
	Name       string
	SliceName  string
	DaoName    string
	Columns    []*ModelColumn
	IsReadOnly bool
	IDType     string
}

type ModelColumn struct {
	*Column
	CamelName       string
	CamelPluralName string
}

func (m *Model) FromTable(t *Table) {
	p := pluralize.NewClient()
	m.Name = strcase.ToCamel(p.Singular(t.Name))
	m.SliceName = p.Plural(m.Name)
	m.DaoName = strcase.ToLowerCamel(m.Name) + "Dao"
	m.Columns = []*ModelColumn{}
	m.IsReadOnly = t.IsReadOnly

	for _, c := range t.Columns {
		fieldName := strcase.ToCamel(c.Name)
		fieldsName := p.Plural(fieldName)
		if strings.HasSuffix(fieldName, "Id") {
			l := len(fieldName)
			fieldName = fieldName[:l-2] + "ID"
			fieldsName = fieldName + "s"
		}
		col := &ModelColumn{
			Column:          c,
			CamelName:       fieldName,
			CamelPluralName: fieldsName,
		}
		if c.Name == "id" {
			m.IDType = string(c.EntityType)
		}
		m.Columns = append(m.Columns, col)
	}
}

func (m *Model) GenerateCode(writer io.Writer, moduleName string) error {
	f := newFile("model")

	entityPackage := fmt.Sprintf("%s/%s", moduleName, EntityPackageName)
	daoPackage := fmt.Sprintf("%s/%s", moduleName, DaoPackageName)
	f.ImportName(entityPackage, EntityPackageName)
	f.ImportName(daoPackage, DaoPackageName)
	f.ImportName(LogLib, "log")
	f.ImportName(ErrorsLib, "errors")
	idType := i(m.IDType)
	idParam := i("id").Add(idType)
	singleEntity := qual(entityPackage, m.Name)
	ptrEntity := ptr().Add(singleEntity)
	entityParam := i("e").Add(ptrEntity)

	instanceName := m.Name + "Instance"
	sliceInstanceName := m.SliceName + "Instance"
	instancePointer := ptr().Id(instanceName)
	sliceInstancePointer := ptr().Id(sliceInstanceName)

	returnNil := rtn().Nil()
	returnErr := rtn(traceErr())

	// define struct
	f.Type().Id(m.Name + "Impl").Struct(
		i(m.DaoName).Qual(daoPackage, m.Name),
	)

	// createInstance
	values := cmap{
		i(m.Name): i("e"),
	}
	if !m.IsReadOnly {
		values[i(m.DaoName)] = i("m").Dot(m.DaoName)
	}
	f.Add(pfn("m", m.Name+"Impl").Id("createInstance").Params(entityParam).Params(instancePointer).Block(
		rtn().Op("&").Id(instanceName).Add(vals(values)),
	)).Line()

	fields := []code{
		ptr().Qual(entityPackage, m.Name),
	}
	if !m.IsReadOnly {
		fields = append(fields, i(m.DaoName).Qual(daoPackage, m.Name))
	}
	f.Type().Id(instanceName).Struct(fields...)

	if !m.IsReadOnly {
		// Save and Delete
		e := i("i").Dot(m.Name)
		for _, name := range []string{"Save", "Delete"} {
			f.Add(pfn("i", instanceName).Id(name).Params().Error().Block(
				ifa(idot("i", m.DaoName), "==", null()).Block(returnNil),
				rtn(qual(ErrorsLib, "Trace").Call(i("i").Dot(m.DaoName).Dot(name).Call(e))),
			)).Line()
		}
	}

	f.Type().Id(sliceInstanceName).Struct(
		i("values").Index().Add(instancePointer),
	).Line()

	// instantiate
	f.Func().Id("New" + sliceInstanceName).Params().Params(sliceInstancePointer).Block(
		rtn(addr().Id(sliceInstanceName).Add(vals(cmap{
			i("values"): idx().Add(instancePointer).Values(),
		}))),
	).Line()

	// Add
	f.Add(pfn("i", sliceInstanceName).Id("Add").Params(i("v").Add(instancePointer)).Block(
		idot("i", "values").Op("=").Append(idot("i", "values"), i("v")),
	)).Line()

	// FindByID
	valueID := idot("v", "ID")
	f.Add(pfn("i", sliceInstanceName).Id("FindByID").Params(idParam).Params(instancePointer).Block(
		forEachV("v", idot("i", "values")).Block(
			ifa(valueID, "==", i("id")).Block(rtn(i("v"))),
		),
		returnNil,
	)).Line()

	// FilterBy
	cbFunc := i("f").Func().Params(instancePointer).Bool()
	f.Add(pfn("i", sliceInstanceName).Id("FilterBy").Params(cbFunc).Params(sliceInstancePointer).Block(
		i("instance").Op(":=").Id("New"+sliceInstanceName).Call(),
		forEachV("v", idot("i", "values")).Block(
			ifb(i("f").Call(i("v"))).Block(
				idot("instance", "Add").Call(i("v")),
			),
		),
		rtn(i("instance")),
	)).Line()

	// Each
	cbFunc = i("f").Func().Params(instancePointer)
	f.Add(pfn("i", sliceInstanceName).Id("Each").Params(cbFunc).Block(
		forEachV("v", idot("i", "values")).Block(
			i("f").Call(i("v")),
		),
	)).Line()

	// EachWithError
	cbFunc = i("f").Func().Params(instancePointer).Error()
	f.Add(pfn("i", sliceInstanceName).Id("EachWithError").Params(cbFunc).Error().Block(
		forEachV("v", idot("i", "values")).Block(
			ifx(i("err").Op(":=").Id("f").Call(i("v")), i("err"), "!=", null()).Block(returnErr),
		),
		returnNil,
	)).Line()

	// First
	f.Add(pfn("i", sliceInstanceName).Id("First").Params().Params(instancePointer).Block(
		ifa(size(idot("i", "values")), "==", lit(0)).Block(returnNil),
		rtn(idot("i", "values").Index(lit(0))),
	)).Line()

	// At
	f.Add(pfn("i", sliceInstanceName).Id("At").Params(i("idx").Id("int")).Params(instancePointer).Block(
		ifa(size(idot("i", "values")), "<", i("idx")).Block(returnNil),
		rtn(idot("i", "values").Index(i("idx"))),
	)).Line()

	// Len
	f.Add(pfn("i", sliceInstanceName).Id("Len").Params().Params(i("int")).Block(
		rtn().Len(i("i").Dot("values")),
	)).Line()

	// IsEmpty
	f.Add(pfn("i", sliceInstanceName).Id("IsEmpty").Params().Params(i("bool")).Block(
		rtn(i("i").Dot("Len").Call().Op("==").Lit(0)),
	)).Line()

	for _, c := range m.Columns {
		var columnType code
		if c.EntityType == TimePtr {
			columnType = ptr().Qual("time", "Time")
		} else {
			columnType = i(string(c.EntityType))
		}
		valueField := idot("v", c.CamelName)

		// FilterByColumn
		name := "FilterBy" + c.CamelName
		f.Add(pfn("i", sliceInstanceName).Id(name).Params(i("c").Add(columnType)).Params(sliceInstancePointer).Block(
			i("s").Op(":=").Id("New"+sliceInstanceName).Call(),
			forEachV("v", idot("i", "values")).Block(
				ifa(valueField, "==", i("c")).Block(
					idot("s", "Add").Call(i("v")),
				),
			),
			rtn(i("s")),
		)).Line()

		// SortByColumn
		valueI := i("s").Dot("values").Index(i("i")).Dot(c.CamelName)
		valueJ := i("s").Dot("values").Index(i("j")).Dot(c.CamelName)
		descCompare := valueI.Clone().Op(">").Add(valueJ)
		ascCompare := valueI.Clone().Op("<").Add(valueJ)
		if c.EntityType == TimePtr {
			descCompare = valueI.Clone().Dot("Before").Call(ptr(valueJ))
			ascCompare = valueI.Clone().Dot("After").Call(ptr(valueJ))
		} else if c.EntityType == Bool {
			descCompare = valueJ
			ascCompare = valueI
		}
		f.Add(pfn("i", sliceInstanceName).Id("SortBy"+c.CamelName).Params(i("isDesc").Id("bool")).Params(sliceInstancePointer).Block(
			i("s").Op(":=").Id("New"+sliceInstanceName).Call(),
			i("s").Dot("values").Op("=").Id("i").Dot("values"),
			qual("sort", "SliceStable").Call(i("s").Dot("values"), fn().Params(i("i"), i("j").Id("int")).Params(i("bool")).Block(
				ifb(i("isDesc")).Block(rtn(descCompare)),
				rtn(ascCompare),
			)),
			rtn(i("s")),
		)).Line()

		// return Column slice
		f.Add(pfn("i", sliceInstanceName).Id(c.CamelPluralName).Params().Params(idx().Add(columnType)).Block(
			i("s").Op(":=").Index().Add(columnType).Values(),
			idot("i", "Each").Call(fn().Params(i("v").Add(instancePointer)).Block(
				i("s").Op("=").Append(i("s"), valueField),
			)),
			rtn(i("s")),
		)).Line()
	}

	if !m.IsReadOnly {
		f.Add(pfn("i", sliceInstanceName).Id("Save").Params().Params(jerr()).Block(
			rtn(i("i").Dot("EachWithError").Call(fn().Params(i("i").Add(ptr(i(instanceName)))).Params(jerr()).Block(
				rtn(traceErr(i("i").Dot("Save").Call())),
			))),
		)).Line()
	}

	return errors.Trace(f.Render(writer))
}
