package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path"
	"text/tabwriter"

	"github.com/BurntSushi/xgb"
	"github.com/BurntSushi/xgb/randr"
)

var (
	command = "table"

	commands = map[string]func(*config, heads){
		"table": table,
		"tabs": tabs,
	}

	flagHeader bool
)

// table runs the 'table' command. The 'table' command outputs a visually
// pleasing and aligned table of the current monitor configuration.
//
// An optional flag, '--header' will show column names.
func table(config *config, heads heads) {
	tabw := tabwriter.NewWriter(os.Stdout, 0, 4, 4, ' ', 0)
	if flagHeader {
		tabw.Write([]byte(
			"Monitor number\t"+
			"Nice output name\t"+
			"Output name\t"+
			"(X, Y)\t"+
			"WidthxHeight\t"+
			"Is primary\n"))
	}
	for i, head := range heads.heads {
		isPrimary := ""
		if head.id == heads.primary.id {
			isPrimary = "primary"
		}
		tabw.Write([]byte(fmt.Sprintf("%d\t%s\t%s\t(%d, %d)\t%dx%d\t%s\n",
			i, config.nice(head.output), head.output,
			head.x, head.y, head.width, head.height, isPrimary)))
	}
	tabw.Flush()
}

// tabs runs the 'tabs' command. The 'tabs' command outputs a tab-delimited
// list of output information. It has precisely the same information as given
// in the 'table' command, but each field is delimited by a tab (i.e., this
// format should be used for parsing).
//
// An optional flag, '--header' will show column names.
func tabs(config *config, heads heads) {
	if flagHeader {
		fmt.Print(
			"Monitor number\t"+
			"Nice output name\t"+
			"Output name\t"+
			"X\tY\t"+
			"Width\tHeight\t"+
			"Is primary\n")
	}
	for i, head := range heads.heads {
		isPrimary := ""
		if head.id == heads.primary.id {
			isPrimary = "primary"
		}
		fmt.Printf("%d\t%s\t%s\t%d\t%d\t%d\t%d\t%s\n",
			i, config.nice(head.output), head.output,
			head.x, head.y, head.width, head.height, isPrimary)
	}
}

func init() {
	log.SetPrefix("[gohead] ")

	flag.BoolVar(&flagHeader, "header", false,
		"If set, column headers will be shown in the 'tabs' and 'table' " +
		"commands.")
	flag.Usage = usage
	flag.Parse()

	if flag.NArg() == 1 {
		command = flag.Arg(0)
	}
}

func usage() {
	fmt.Fprintf(os.Stderr,
		"Usage: %s [ table | tabs ]\n",
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

