package config

import (
	"fmt"
	"io/ioutil"
	"flag"
	"log"
	"path"
	"encoding/json"
	"gopkg.in/yaml.v3"
)

// Wrapper wraps a config struct to enable flag.Var() to load configs from file.
type Wrapper struct {
	obj interface{}
}

func FromYamlFile(filename string, obj interface{}) error {
	bytes, err := ioutil.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("Could not read YAML config file at %s: %v", filename, err)
	}
	if err := yaml.Unmarshal(bytes, obj); err != nil {
		return fmt.Errorf("Could not load YAML config file at %s: %v", filename, err)
	}
	return nil
}

func FromJsonFile(filename string, obj interface{}) error {
	bytes, err := ioutil.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("Could not read JSON config file at %s: %v", filename, err)
	}
	if err := json.Unmarshal(bytes, obj); err != nil {
		return fmt.Errorf("Could not load JSON config file at %s: %v", filename, err)
	}
	return nil
}

func FromFile(filename string, obj interface{}) error {
	ext := path.Ext(filename)
	switch ext {
	case ".yaml", ".yml":
		return FromYamlFile(filename, obj)
	case ".json":
		return FromJsonFile(filename, obj)
	default:
		return fmt.Errorf("Could not recognize format of config file %s. Unknown extension '%s'.", filename, ext)
	}
}

func FromFileOrDie(filename string, obj interface{}) {
	if err := FromFile(filename, obj); err != nil {
		panic(err)
	}
}

func FromArgs(args []string, obj interface{}) error {
	var filename string
	flagset := flag.NewFlagSet("config", flag.ExitOnError)
	flagset.StringVar(&filename, "config", "", "Config file. Supports YAML, JSON.")
	flagset.Parse(args)
	if filename != "" {
		log.Printf("Reading configuration from file %s.", filename)
		return FromFile(filename, obj)
	} else {
		return nil
	}
}

func FromArgsOrDie(args []string, obj interface{}) error {
	if err := FromArgs(args, obj); err != nil {
		log.Fatal(err)	
	}
	return nil
}

// Flag creates a --config flag to load configuration from file.
func Flag(obj interface{}) {
	flag.Var(&Wrapper{obj}, "config", "Config file. Supports YAML, JSON.")
}

func (w Wrapper) Set(f string) error {
	return FromFile(f, w.obj)
}

func (w Wrapper) String() string {
	return fmt.Sprintf("%v", w.obj)
}
