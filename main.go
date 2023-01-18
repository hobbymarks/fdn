/*
Package main
Copyright © 2022 hobbymarks ihobbymarks@gmail.com
*/
package main

import (
	"github.com/hobbymarks/fdn/cmd"
	log "github.com/sirupsen/logrus"
)

// Debug debug flag
var Debug bool = true

func main() {
	if Debug {
		log.SetReportCaller(true)
	} else {
		log.SetReportCaller(false)
	}
	// log.SetLevel(log.TraceLevel)

	cmd.Execute()
}
