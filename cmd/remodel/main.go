package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/juju/errors"
	"github.com/yuki-eto/remodel"
)

func main() {
	if err := run(); err != nil {
		log.Fatalf("err: %+v", err)
	}
}

func run() error {
	var (
		rootDir    string
		moduleName string
		isProtoc   bool
		isJSON     bool
	)
	flag.StringVar(&rootDir, "root", "", "root directory of project")
	flag.StringVar(&moduleName, "module", "", "module name of project")
	flag.BoolVar(&isProtoc, "proto", false, "necessary protocol buffers schema for entity")
	flag.BoolVar(&isJSON, "json", false, "necessary json output")
	flag.Parse()

	if rootDir == "" {
		flag.Usage()
		return nil
	}

	mode := flag.Arg(0)
	if mode == "yaml" {
		s := &remodel.Tables{}
		return errors.Trace(s.Output(rootDir))
	}

	ts := &remodel.Tables{}
	if err := ts.Load(rootDir); err != nil {
		return errors.Trace(err)
	}

	switch mode {
	case "entity":
		s := ts.Entities()
		return errors.Trace(s.Output(rootDir, isJSON, isJSON))
	case "dao":
		if moduleName == "" {
			flag.Usage()
			return nil
		}
		s := ts.Daos()
		return errors.Trace(s.Output(rootDir, moduleName))
	case "model":
		if moduleName == "" {
			flag.Usage()
			return nil
		}
		s := ts.Models()
		return errors.Trace(s.Output(rootDir, moduleName))
	default:
		fmt.Println("please input mode: [yaml|entity|dao|model]")
		return nil
	}
}
