package main

import (
	"log"
	"os"
	"path"

	ini "github.com/glacjay/goini"
)

var (
	xdgfile = path.Join(os.Getenv("XDG_CONFIG_HOME"), "gohead", "config.ini")
	myfile  = path.Join(os.Getenv("HOME"), ".config", "gohead", "config.ini")
)

type config struct {
	// outputs is a map from output names (i.e., "LVDS1") to nice names
	// specified in a config file (i.e., "laptop").
	outputs map[string]string
}

// newConfig checks for a ini file in XDG_CONFIG_HOME/gohead/config.ini, then
// looks for $HOME/.config/gohead/config.ini. If neither one can be found,
// a default (blank) configuration is used.
func newConfig() *config {
	conf := &config{
		outputs: make(map[string]string, 10),
	}

	dict, err := ini.Load(xdgfile)
	if err != nil {
		if os.IsNotExist(err) && xdgfile != myfile {
			dict, err = ini.Load(myfile)
			if err != nil {
				log.Printf("Neither '%s' nor '%s' could be read: %s.",
					xdgfile, myfile, err)
				return conf
			}
		} else {
			log.Printf("There was an error when trying to load '%s': %s.",
				xdgfile, err)
			return conf
		}
	}

	for _, section := range dict.GetSections() {
		switch section {
		case "monitors":
			for niceName := range dict[section] {
				if output, ok := dict.GetString(section, niceName); ok {
					conf.outputs[output] = niceName
				}
			}
		case "":
		default:
			log.Printf("I don't know what to do with section '%s'.", section)
		}
	}
	return conf
}

func (c *config) nice(output string) string {
	if niceName, ok := c.outputs[output]; ok {
		return niceName
	}
	return output
}
