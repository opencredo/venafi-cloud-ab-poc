package config

import (
	"flag"
	"fmt"
	"os"
	"strings"
)

var envPrefix string

type configValue struct {
	value        string
	loc          *string
	name         string
	defaultValue string
	usage        string
}

var configValues []*configValue

func Prefix(p string) {
	envPrefix = p
}

func StringVar(loc *string, name string, value string, usage string) {
	v := configValue{loc: loc, name: name, defaultValue: value, usage: fmt.Sprintf("%s (env: %s%s)", usage, envPrefix, strings.ToUpper(name))}
	configValues = append(configValues, &v)

	flag.StringVar(&(v.value), v.name, "", v.usage)
}

func Parse() {
	flag.Parse()

	for _, v := range configValues {
		if len(v.value) > 0 {
			*v.loc = v.value
			continue
		}

		envName := fmt.Sprint(envPrefix, strings.ToUpper(v.name))
		envValue := os.Getenv(envName)
		if len(envValue) > 0 {
			*v.loc = envValue
			continue
		}

		*v.loc = v.defaultValue
	}
}
