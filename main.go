/*
Copyright © 2022 hobbymarks ihobbymarks@gmail.com
*/
package main

import (
	"github.com/hobbymarks/fdn/cmd"
	log "github.com/sirupsen/logrus"
)

func main() {
	log.SetReportCaller(true)
	log.SetLevel(log.TraceLevel)

	cmd.Execute()
}
