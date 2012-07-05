package main

import (
	"flag"
	"fmt"
	"os"
	"text/tabwriter"
)

// table runs the 'table' command. The 'table' command outputs a visually
// pleasing and aligned table of the current monitor configuration.
//
// An optional flag, '--header' will show column names.
func table(config *config, heads heads) {
	tabw := tabwriter.NewWriter(os.Stdout, 0, 4, 4, ' ', 0)
	if flagHeader {
		tabw.Write([]byte(
			"Monitor number\t" +
				"Nice output name\t" +
				"Output name\t" +
				"(X, Y)\t" +
				"WidthxHeight\t" +
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
			"Monitor number\t" +
				"Nice output name\t" +
				"Output name\t" +
				"X\tY\t" +
				"Width\tHeight\t" +
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

// query returns the monitor geometry for the given output name.
// The output name can either be a nice name specified in the configuration
// file, or the actual RandR output name. The special output name 'primary'
// will return the geometry for the primary monitor.
func query(config *config, heads heads) {
	if flag.NArg() < 2 {
		fmt.Fprintln(os.Stderr, "The 'query' command expects a head name, but "+
			"didn't find one.")
		usage()
	}
	headName := flag.Arg(1)
	showHead := func(hd head) {
		fmt.Printf("%s\t%d\t%d\t%d\t%d\n", config.nice(hd.output),
			hd.x, hd.y, hd.width, hd.height)
	}

	if headName == "primary" {
		showHead(*heads.primary)
		return
	}
	for _, hd := range heads.heads {
		if hd.output == headName || config.nice(hd.output) == headName {
			showHead(hd)
			return
		}
	}

	fmt.Fprintf(os.Stderr, "The 'query' command could not find a head "+
		"matching the name '%s'.\n", headName)
	os.Exit(1)
}
