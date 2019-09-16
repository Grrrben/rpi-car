package rpi_car

import (
	"github.com/golang/glog"
	"github.com/grrrben/gpio"
	"time"
)

// used pins in BCM numbering
const pinLedFront = 10
const pinLedBack = 9

const pinSensorFrontTrigger = 8
const pinSensorFrontEcho = 7
const pinSensorBackTrigger = 20
const pinSensorBackEcho = 21

const pinMotorPlus = 17
const pinMotorMinus = 18
const pinMotorEnable = 27

type lighting struct {
	front *gpio.Led
	back  *gpio.Led
}

type distanceSensors struct {
	front *gpio.HCSR04
	back  *gpio.HCSR04
}

// a wrapper to control all different gpio of a car
type car struct {
	lights  lighting
	sensors distanceSensors
	// propulsion
	motor *gpio.Motor
}

func NewCar() *car {
	c := new(car)

	c.lights.front = gpio.NewLed(pinLedFront)
	c.lights.back = gpio.NewLed(pinLedBack)
	c.sensors.front = gpio.NewHCSR04(pinSensorFrontTrigger, pinSensorFrontEcho)
	c.sensors.back = gpio.NewHCSR04(pinSensorBackTrigger, pinSensorBackEcho)
	c.motor = gpio.NewMotor(pinMotorPlus, pinMotorMinus, pinMotorEnable)

	return c
}

func (c *car) Init() {
	glog.Info("Car initiated")

	c.lights.front.Blink()
	c.lights.back.Blink()

	c.drive()
}

func (c *car) drive() {
	front := make(chan float64, 3)
	back := make(chan float64, 3)

	go func() {
		for {
			front <- c.sensors.front.Measure()
			time.Sleep(time.Second / 2)
		}
	}()

	go func() {
		for {
			back <- c.sensors.back.Measure()
			time.Sleep(time.Second / 2)
		}
	}()

	// only take action if both readings are known
	f := false
	b := false

	// caching last distances for taking measurements
	var lastFront, lastBack float64

	for {
		select {
		case cmFront := <-front:
			f = true
			lastFront = cmFront
			if f && b {
				diff := lastBack - lastFront
				if diff < 0 {
					c.forwards()
				}
			}
		case cmBack := <-back:
			b = true
			lastBack = cmBack
			if f && b {
				diff := lastFront - lastBack
				if diff < 0 {
					c.backwards()
				}
			}
		}
	}
}

func (c *car) forwards() {
	c.motor.Forwards()
	c.lights.front.Blink()
	glog.Info("Moving forwards")
}

func (c *car) backwards() {
	c.motor.Backwards()
	c.lights.back.Blink()
	glog.Info("Moving backwards")
}

func (c *car) stop() {
	c.motor.Stop()
	glog.Info("Stopped moving")

}
