package remodel

import (
	"fmt"
	"io"
	"strings"

	"github.com/gertd/go-pluralize"
	"github.com/iancoleman/strcase"
	"github.com/juju/errors"
)

type EntityType string

type Entity struct {
	Name          string
	SliceName     string
	TableName     string
	Fields        []*Field
	IsReadOnly    bool
	EntityPackage string
}

type Field struct {
	Name       string
	ColumnName string
	FieldType  EntityType
}

func (e *Entity) FromTable(t *Table) {
	p := pluralize.NewClient()
	e.Name = strcase.ToCamel(p.Singular(t.Name))
	e.SliceName = strcase.ToCamel(t.Name)
	e.TableName = t.Name
	e.IsReadOnly = t.IsReadOnly

	for _, c := range t.Columns {
		fieldName := strcase.ToCamel(c.Name)
		if strings.HasSuffix(fieldName, "Id") {
			l := len(fieldName)
			fieldName = fieldName[:l-2] + "ID"
		}
		f := &Field{
			Name:       fieldName,
			ColumnName: c.Name,
			FieldType:  c.EntityType,
		}
		e.Fields = append(e.Fields, f)
	}
}

func (e *Entity) GenerateCode(writer io.Writer) error {
	f := newFile("entity")
	f.ImportName(ErrorsLib, "errors")
	f.ImportName(RapidashLib, "rapidash")

	var fields []code
	for _, field := range e.Fields {
		fields = append(fields, e.fieldToCode(field))
	}
	// define struct
	f.Type().Id(e.Name).Struct(fields...).Line()
	f.Type().Id(e.SliceName).Index().Op("*").Id(e.Name).Line()

	var (
		decodeCodes []code
	)
	encodeCodes := []code{
		ifa(i("e").Dot("ID"), "!=", lit(0)).Block(
			i("enc").Dot("Uint64").Call(lit("id"), i("e").Dot("ID")),
		),
	}
	structCodes := []code{
		i("s").Op(":=").Qual(RapidashLib, "NewStruct").Call(lit(e.TableName)),
	}
	for _, field := range e.Fields {
		fieldName := field.Name
		decodeCode := i("e").Dot(fieldName).Op("=")
		fieldType := strcase.ToCamel(string(field.FieldType))
		structType := fieldType
		if field.FieldType == TimePtr {
			fieldType = "TimePtr"
			structType = "Time"
		}
		decodeCode.Id("dec").Dot(fieldType).Call(lit(field.ColumnName))
		encodeCode := i("enc").Dot(fieldType).Call(lit(field.ColumnName), i("e").Dot(field.Name))
		encodeCodes = append(encodeCodes, encodeCode)
		decodeCodes = append(decodeCodes, decodeCode)
		structCodes = append(structCodes, i("s").Dot("Field"+structType).Call(lit(field.ColumnName)))
	}
	encodeCodes = append(encodeCodes, rtn(i("enc").Dot("Error").Call()))
	decodeCodes = append(decodeCodes, rtn(i("dec").Dot("Error").Call()))
	structCodes = append(structCodes, rtn(i("s")))

	if !e.IsReadOnly {
		f.Add(pfn("e", e.Name).Id("EncodeRapidash").Params(i("enc").Qual(RapidashLib, "Encoder")).Error().Block(
			encodeCodes...,
		)).Line()
		f.Add(pfn("e", e.SliceName).Id("EncodeRapidash").Params(i("enc").Qual(RapidashLib, "Encoder")).Error()).Block(
			forEachV("v", ptr(i("e"))).Block(
				ifxErr(i("v").Dot("EncodeRapidash").Call(i("enc").Dot("New").Call())).Block(
					rtn(traceErr()),
				),
			),
			rtn(null()),
		).Line()
	}
	f.Add(pfn("e", e.Name).Id("DecodeRapidash").Params(i("dec").Qual(RapidashLib, "Decoder")).Error().Block(decodeCodes...)).Line()
	f.Add(pfn("e", e.SliceName).Id("DecodeRapidash").Params(i("dec").Qual(RapidashLib, "Decoder")).Error().Block(
		i("count").Op(":=").Id("dec").Dot("Len").Call(),
		ptr(i("e")).Op("=").Make(idx().Add(ptr(i(e.Name))), i("count")),
		forItr("i", lit(0), "<", i("count")).Block(
			jvar("v").Id(e.Name),
			ifxErr(i("v").Dot("DecodeRapidash").Call(i("dec").Dot("At").Call(i("i")))).Block(
				rtn(traceErr()),
			),
			op("(").Add(ptr(i("e"))).Op(")").Index(i("i")).Op("=").Add(addr(i("v"))),
		),
		rtn(null()),
	)).Line()
	f.Add(pfn("e", e.Name).Id("Struct").Params().Params(ptr().Add(qual(RapidashLib, "Struct"))).Block(structCodes...)).Line()

	return errors.Trace(f.Render(writer))
}

func (f *Field) typeToCode() code {
	if f.FieldType == TimePtr {
		return ptr().Qual("time", "Time")
	}
	return i(string(f.FieldType))
}

func (e *Entity) fieldToCode(f *Field) code {
	jf := i(f.Name).Add(f.typeToCode())
	tags := map[string]string{}
	if e.IsReadOnly {
		tags["csv"] = strcase.ToCamel(f.ColumnName)
	}
	jf.Tag(tags)
	return jf
}

func (e *Entity) fieldToROCode(f *Field) code {
	return i(strcase.ToLowerCamel(f.ColumnName)).Add(f.typeToCode())
}

func (e *Entity) fieldToGetMethodCode(f *Field) code {
	return fn().Params(i("e").Add(i(e.Name))).Id(f.Name).Params().Params(f.typeToCode()).Block(
		rtn(i("e").Dot(strcase.ToLowerCamel(f.ColumnName))),
	)
}

type Entities []*Entity

func (e *Entities) GenerateStructableCode(writer io.Writer, moduleName string) error {
	f := newFile("entity")

	entityPackage := fmt.Sprintf("%s/entity", moduleName)
	f.ImportName(entityPackage, "entity")
	f.ImportName(RapidashLib, "rapidash")

	f.Type().Id("Structable").Interface(
		i("Struct").Call().Add(ptr(qual(RapidashLib, "Struct"))),
	)

	tables := cmap{}
	roTables := cmap{}
	for _, v := range *e {
		if v.IsReadOnly {
			roTables[lit(v.TableName)] = op("new").Call(qual(entityPackage, v.Name))
		} else {
			tables[lit(v.TableName)] = op("new").Call(qual(entityPackage, v.Name))
		}
	}

	f.Func().Id("Structables").Call().Params(jmap(str()).Id("Structable")).Block(
		rtn().Map(str()).Id("Structable").Add(vals(tables)),
	)
	f.Func().Id("ReadOnlyStructables").Call().Params(jmap(str()).Id("Structable")).Block(
		rtn().Map(str()).Id("Structable").Add(vals(roTables)),
	)

	return errors.Trace(f.Render(writer))
}
