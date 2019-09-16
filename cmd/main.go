package main

import (
	"flag"
	"fmt"
	"github.com/grrrben/glog"
	"github.com/grrrben/rpi-car"
	"github.com/stianeikeland/go-rpio"
	"os"
)

func main() {

	flag.Parse()

	glog.SetLogLevel(glog.Log_level_info)
	glog.SetOutput(os.Stdout)

	// Open and map memory to access gpio, check for errors
	if err := rpio.Open(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// Unmap gpio memory when done
	defer rpio.Close()

	car := rpi_car.NewCar()
	car.Init()

	// wait for everything to finish
	forever := make(chan bool)
	<-forever
}
