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
	// A map from nice name (e.g., "laptop") to particular settings for
	// that head.
	headConfigs map[string]headConfig
}

type headConfig struct {
	// An RandR mode for this head.
	mode string
}

// newConfig checks for a ini file in XDG_CONFIG_HOME/gohead/config.ini, then
// looks for $HOME/.config/gohead/config.ini. If neither one can be found,
// a default (blank) configuration is used.
func newConfig() *config {
	var dict ini.Dict
	var err error

	conf := &config{
		outputs:     make(map[string]string),
		headConfigs: make(map[string]headConfig),
	}

	if len(flagConfig) > 0 {
		dict, err = ini.Load(flagConfig)
		if err != nil {
			log.Printf("There was an error when trying to load '%s': %s.",
				flagConfig, err)
			return conf
		}
	} else {
		dict, err = ini.Load(xdgfile)
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
			hconfig := headConfig{}
			for key := range dict[section] {
				switch key {
				case "mode":
					if mode, ok := dict.GetString(section, key); ok {
						hconfig.mode = mode
					}
				default:
					log.Printf(
						"unknown head config key '%s.%s'",
						section,
						key,
					)
				}
			}
			conf.headConfigs[section] = hconfig
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
