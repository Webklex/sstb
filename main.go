package main

import (
	"./app"
	"./utils/log"
	"flag"
	"fmt"
	"os"
)

var buildNumber string
var buildVersion string

func main() {
	ac := app.DefaultConfig()
	lc := log.DefaultConfig()

	ac.AddFlags(flag.CommandLine)
	lc.AddFlags(flag.CommandLine)

	sv := flag.Bool("version", false, "Show version and exit")
	flag.Parse()

	ac.Build = app.Build{
		Number:  buildNumber,
		Version: buildVersion,
	}

	if *sv {
		fmt.Printf("Version: %s\n", ac.Build.Version)
		fmt.Printf("Build number: %s\n", ac.Build.Number)
		return
	}

	lc.Load(lc.File)
	log.Log = lc
	log.Log.Init()

	ac.Load(ac.File)

	_ = os.Setenv("TZ", ac.Timezone)

	a := app.NewApp(ac)
	a.Start()
}
