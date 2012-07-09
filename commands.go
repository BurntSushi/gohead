package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
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

	// For each head name specified on the command line, verify that it is:
	// 1) Not 'primary'.
	// 2) A valid head name that is connected.
	// 3) Not specified more than once.
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

	// Now gobble up ALL of the heads that *aren't* in toenable. This includes
	// heads that are disconnected! The reason is that a head can be
	// disconnected but still "enabled" in xrandr's eyes, leading to interesting
	// and undesirable states.
	todisable := make([]string, 0)
	for _, head := range heads.heads {
		if !icontains(head.output, toenable) {
			todisable = append(todisable, head.output)
		}
	}
	for _, name := range heads.off {
		if !icontains(name, toenable) {
			todisable = append(todisable, name)
		}
	}
	for _, name := range heads.disconnected {
		if !icontains(name, toenable) {
			todisable = append(todisable, name)
		}
	}

	// Construct the xrandr command, print it, then execute it. Echo its output.
	args := xrandrArgs(toenable, todisable, flagVertical)
	fmt.Printf("xrandr %s\n", strings.Join(args, " "))

	if flagTest {
		return
	}

	cmd := exec.Command("xrandr", args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Println(string(out))
		fmt.Println(err)
		os.Exit(1)
	} else {
		fmt.Print(string(out))
	}
}

func xrandrArgs(toenable, todisable []string, vertical bool) []string {
	args := make([]string, 0, len(toenable)*5)
	first := true
	for i, name := range toenable {
		switch {
		case first:
			args = append(args, "--output", name, "--auto")
			first = false
		case vertical:
			args = append(args, "--output", name, "--auto",
				"--below", toenable[i-1])
		default:
			args = append(args, "--output", name, "--auto",
				"--right-of", toenable[i-1])
		}
	}
	for _, name := range todisable {
		args = append(args, "--output", name, "--off")
	}
	return args
}

// primary sets the specified *connected* head to be the primary monitor.
func primary(config *config, heads heads) {
	if flag.NArg() != 2 {
		fmt.Fprintf(os.Stderr, "The 'primary' command expects one "+
			"head name, but found %d head names.\n\n", flag.NArg()-1)
		flag.Usage()
	}

	xname := heads.connectedRandrName(config, flag.Arg(1))
	if len(xname) == 0 {
		fmt.Fprintf(os.Stderr, "The head name '%s' does not "+
			"refer to a connected monitor.\n", flag.Arg(1))
		os.Exit(1)
	}

	args := []string{"--output", xname, "--primary"}
	fmt.Println("xrandr", strings.Join(args, " "))

	if flagTest {
		return
	}

	cmd := exec.Command("xrandr", args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Println(string(out))
		fmt.Println(err)
		os.Exit(1)
	} else {
		fmt.Print(string(out))
	}
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
