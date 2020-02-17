package remodel

import (
	"github.com/dave/jennifer/jen"
)

type code = jen.Code
type statement = jen.Statement
type cmap map[code]code

func newFile(packageName string) *jen.File {
	return jen.NewFile(packageName)
}

func qual(path, name string) *statement {
	return jen.Qual(path, name)
}

func op(o string) *statement {
	return jen.Op(o)
}

func ptr(codes ...code) *statement {
	return op("*").Add(codes...)
}

func addr(codes ...code) *statement {
	return op("&").Add(codes...)
}

func i(name string) *statement {
	return jen.Id(name)
}

func fn() *statement {
	return jen.Func()
}

func pfn(receiver, id string) *statement {
	return jen.Func().Params(i(receiver).Op("*").Id(id))
}

func forItr(variable string, start code, op string, last code) *statement {
	v := i(variable)
	return jen.For().Add(v).Op(":=").Add(start).Op(";").Add(v).Op(op).Add(last).Op(";").Add(v).Op("++")
}

func forEach(k, v string, slice code) *statement {
	var key code
	if k == "_" {
		key = op("_")
	} else {
		key = i(k)
	}
	return jen.For().List(key, i(v)).Op(":=").Range().Add(slice)
}

func forEachK(k string, slice code) *statement {
	return jen.For().Id(k).Op(":=").Range().Add(slice)
}

func forEachV(v string, slice code) *statement {
	return jen.For().List(op("_"), i(v)).Op(":=").Range().Add(slice)
}

func rtn(results ...code) *statement {
	return jen.Return(results...)
}

func traceErr(codes ...code) *statement {
	q := qual(ErrorsLib, "Trace")
	if len(codes) > 0 {
		return q.Call(codes[0])
	}
	return q.Call(i("err"))
}

func ifa(a code, op string, b code) *statement {
	return jen.If().Add(a).Op(op).Add(b)
}

func ifb(b code) *statement {
	return jen.If(b)
}

func ifx(x, a code, op string, b code) *statement {
	return jen.If().Add(x).Op(";").Add(a).Op(op).Add(b)
}

func ifErr() *statement {
	return jen.If().Id("err").Op("!=").Nil()
}

func ifxErr(x code) *statement {
	return ifxErrCustom(i("err").Op(":=").Add(x))
}

func ifxBool(x, a code) *statement {
	return jen.If().Add(x).Op(";").Add(a)
}

func ifxErrCustom(x code) *statement {
	return ifx(x, i("err"), "!=", null())
}

func lit(v interface{}) *statement {
	return jen.Lit(v)
}

func idot(id string, fields ...string) *statement {
	stmt := i(id)
	for _, f := range fields {
		stmt.Dot(f)
	}
	return stmt
}

func idx(items ...code) *statement {
	return jen.Index(items...)
}

func size(code code) *statement {
	return jen.Len(code)
}

func null() *statement {
	return jen.Nil()
}

func vals(m cmap) *statement {
	dict := jen.Dict{}
	for k, v := range m {
		dict[k] = v
	}
	return jen.Values(dict)
}

func list(codes ...code) *statement {
	return jen.List(codes...)
}

func jerr() *statement {
	return jen.Error()
}

func jvar(name string) *statement {
	return jen.Var().Id(name)
}

func u64() *statement {
	return jen.Uint64()
}

func jgo() *statement {
	return jen.Go()
}

func str() *statement {
	return jen.String()
}

func jcontinue() *statement {
	return jen.Continue()
}

func bools(b bool) *statement {
	if b {
		return jen.True()
	}
	return jen.False()
}

func jmap(code code) *statement {
	return jen.Map(code)
}
