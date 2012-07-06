package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path"

	"github.com/BurntSushi/xgb"
	"github.com/BurntSushi/xgb/randr"
)

var (
	command = "set"

	commands = map[string]func(*config, heads){
		"set":       set,
		"table":     table,
		"tabs":      tabs,
		"list":      list,
		"query":     query,
		"connected": connected,
		"all":       all,
	}

	flagConfig string
	flagHeader bool
)

func init() {
	log.SetPrefix("[gohead] ")

	flag.StringVar(&flagConfig, "config", "",
		"Override the default location of the config file.")
	flag.BoolVar(&flagHeader, "header", false,
		"If set, column headers will be shown in the 'tabs' and 'table' "+
			"commands.")
	flag.Usage = usage
	flag.Parse()

	if flag.NArg() >= 1 {
		command = flag.Arg(0)
	}
}

func usage() {
	fmt.Fprintf(os.Stderr,
		"Usage: %s [ set head-name [ head-name ... ] | table | tabs | list "+
			"| query [ head-name | primary ] ]\n",
		path.Base(os.Args[0]))
	flag.PrintDefaults()
	os.Exit(1)
}

func main() {
	X, err := xgb.NewConn()
	if err != nil {
		log.Fatal(err)
	}

	err = randr.Init(X)
	if err != nil {
		log.Fatal(err)
	}

	conf := newConfig()
	hds := newHeads(X)

	if f, ok := commands[command]; ok {
		f(conf, hds)
	} else {
		fmt.Fprintf(os.Stderr, "Unknown command '%s'.\n", command)
		usage()
	}
	os.Exit(0)
}
