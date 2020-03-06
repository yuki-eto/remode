package remodel

import (
	"io/ioutil"
	"os"

	"github.com/juju/errors"
	"golang.org/x/tools/imports"
)

func applyGoimports(codePath string) error {
	importOpts := &imports.Options{
		TabWidth:  4,
		TabIndent: true,
		Comments:  true,
		Fragment:  true,
	}
	b, err := imports.Process(codePath, nil, importOpts)
	if err != nil {
		return errors.Trace(err)
	}
	if err := ioutil.WriteFile(codePath, b, os.ModePerm); err != nil {
		return errors.Trace(err)
	}
	return nil
}
