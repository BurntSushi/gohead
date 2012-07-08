package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path"
	"sort"

	"github.com/BurntSushi/xgb"
	"github.com/BurntSushi/xgb/randr"
)

type commandInfo struct {
	f     func(*config, heads)
	usage string
}

var (
	command = "set"

	commands = map[string]commandInfo{
		"set":       {set, "set HEAD-NAME [ HEAD-NAME ... ]"},
		"table":     {table, "table"},
		"tabs":      {tabs, "tabs"},
		"list":      {list, "list"},
		"query":     {query, "query [ HEAD-NAME | primary ]"},
		"connected": {connected, "connected"},
		"all":       {all, "all"},
	}

	flagBaseline string
	flagConfig   string
	flagHeader   bool
	flagVertical bool
)

func init() {
	log.SetPrefix("[gohead] ")

	flag.StringVar(&flagConfig, "config", "",
		"Override the default location of the config file.")
	flag.BoolVar(&flagHeader, "header", false,
		"If set, column headers will be shown in the 'tabs' and 'table' "+
			"commands.")
	flag.BoolVar(&flagVertical, "vertical", false,
		"If set, monitors will be aligned vertically instead of horizontally.")
	flag.Usage = usage(commands)
	flag.Parse()

	if flag.NArg() >= 1 {
		command = flag.Arg(0)
	}
}

func usage(commands map[string]commandInfo) func() {
	return func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [ flags ] [ command ]\n",
			path.Base(os.Args[0]))
		fmt.Fprint(os.Stderr, "Where 'command' (default: 'set') is one of:\n")
		printCommands(commands)
		flag.PrintDefaults()
		os.Exit(1)
	}
}

func printCommands(commands map[string]commandInfo) {
	lines := make([]string, 0, len(commands))
	for _, cmd := range commands {
		lines = append(lines, cmd.usage)
	}

	sort.Sort(sort.StringSlice(lines))
	for _, line := range lines {
		fmt.Fprintf(os.Stderr, "\t%s\n", line)
	}
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

	if cmdInfo, ok := commands[command]; ok {
		cmdInfo.f(conf, hds)
	} else {
		fmt.Fprintf(os.Stderr, "Unknown command '%s'.\n", command)
		flag.Usage()
	}
	os.Exit(0)
}
