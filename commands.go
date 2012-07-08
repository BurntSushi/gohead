package main

import (
	"flag"
	"fmt"
	"os"
	"sort"
	"strings"
	"text/tabwriter"
)

// set takes a list of output names and issues the appropriate `xrandr` command
// to set them up.
func set(config *config, heads heads) {
	if flag.NArg() < 2 {
		fmt.Fprint(os.Stderr, "The 'set' command expects at least one "+
			"head name, but didn't find any.\n\n")
		flag.Usage()
	}
	toenable := make([]string, flag.NArg()-1)
	for i, headName := range flag.Args()[1:] {
		if headName == "primary" {
			fmt.Fprintf(os.Stderr, "The 'set' command requires specific "+
				"head names, which does not include 'primary'.\n")
			os.Exit(1)
		}

		xname := heads.connectedRandrName(config, headName)
		if len(xname) == 0 {
			fmt.Fprintf(os.Stderr, "The head name '%s' does not "+
				"refer to a connected monitor.\n", headName)
			os.Exit(1)
		}
		if icontains(xname, toenable) {
			fmt.Fprintf(os.Stderr, "The head name '%s' (nice: '%s') "+
				"was specified twice, which is not allowed.\n",
				xname, config.nice(xname))
			os.Exit(1)
		}
		toenable[i] = xname
	}
	if len(toenable) == 0 {
		panic("unreachable")
	}
	fmt.Println(xrandr(toenable, flagVertical, flagBaseline))
}

func xrandr(headNames []string, vertical bool, baseline string) string {
	if len(baseline) > 0 && !icontains(baseline, headNames) {
		fmt.Fprintf(os.Stderr, "The baseline monitor specified, '%s', is "+
			"not in the list of monitors to enable.", baseline)
		os.Exit(1)
	}
	outputs := make([]string, len(headNames))
	first := true
	for i, name := range headNames {
		switch {
		case first:
			outputs[i] = fmt.Sprintf("--output %s --auto", name)
			first = false
		case vertical:
			outputs[i] = fmt.Sprintf("--output %s --auto --below %s",
				name, headNames[i-1])
		default:
			outputs[i] = fmt.Sprintf("--output %s --auto --right-of %s",
				name, headNames[i-1])
		}
	}
	return fmt.Sprintf("xrandr %s", strings.Join(outputs, " "))
}

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

// list prints the output name of each enabled monitor on a new line.
func list(config *config, heads heads) {
	for _, hd := range heads.heads {
		fmt.Println(config.nice(hd.output))
	}
}

// query returns the monitor geometry for the given output name.
// The output name can either be a nice name specified in the configuration
// file, or the actual RandR output name. The special output name 'primary'
// will return the geometry for the primary monitor.
func query(config *config, heads heads) {
	if flag.NArg() < 2 {
		fmt.Fprint(os.Stderr, "The 'query' command expects a head name, but "+
			"didn't find one.\n\n")
		flag.Usage()
	}

	headName := flag.Arg(1)
	if found := heads.findActive(config, headName); found != nil {
		fmt.Printf("%s\t%d\t%d\t%d\t%d\n", config.nice(found.output),
			found.x, found.y, found.width, found.height)
		return
	}
	if found := heads.findOff(config, headName); len(found) > 0 {
		fmt.Printf("%s disabled\n", config.nice(found))
	}

	fmt.Fprintf(os.Stderr, "The 'query' command could not find a head "+
		"matching the name '%s'.\n", headName)
	os.Exit(1)
}

// connected returns a list of all outputs that are currently connected. It
// includes outputs that are disabled.
func connected(config *config, heads heads) {
	outputs := make([]string, 0, len(heads.heads)+len(heads.off))
	for _, hd := range heads.heads {
		outputs = append(outputs, fmt.Sprintf("%s (%d, %d) %dx%d",
			config.nice(hd.output), hd.x, hd.y, hd.width, hd.height))
	}
	for _, name := range heads.off {
		outputs = append(outputs, fmt.Sprintf("%s disabled", config.nice(name)))
	}
	sort.Sort(sort.StringSlice(outputs))
	for _, output := range outputs {
		fmt.Println(output)
	}
}

// all returns a list of all outputs available. It
// includes outputs that are disabled or disconnected.
func all(config *config, heads heads) {
	outputs := make([]string, 0,
		len(heads.heads)+len(heads.off)+len(heads.disconnected))
	for _, hd := range heads.heads {
		outputs = append(outputs, fmt.Sprintf("%s (%d, %d) %dx%d",
			config.nice(hd.output), hd.x, hd.y, hd.width, hd.height))
	}
	for _, name := range heads.off {
		outputs = append(outputs, fmt.Sprintf("%s disabled", config.nice(name)))
	}
	for _, name := range heads.disconnected {
		outputs = append(outputs, fmt.Sprintf("%s disconnected",
			config.nice(name)))
	}
	sort.Sort(sort.StringSlice(outputs))
	for _, output := range outputs {
		fmt.Println(output)
	}
}
