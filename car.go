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

const pinMotorLeftPlus = 17
const pinMotorLeftMinus = 18
const pinMotorLeftEnable = 27

const pinMotorRightPlus = 22
const pinMotorRightMinus = 23
const pinMotorRightEnable = 24

type lighting struct {
	front *gpio.Led
	back  *gpio.Led
}

type propulsion struct {
	left  *gpio.Motor
	right *gpio.Motor
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
	motors propulsion
}

func NewCar() *car {
	c := new(car)

	c.lights.front = gpio.NewLed(pinLedFront)
	c.lights.back = gpio.NewLed(pinLedBack)
	c.sensors.front = gpio.NewHCSR04(pinSensorFrontTrigger, pinSensorFrontEcho)
	c.sensors.back = gpio.NewHCSR04(pinSensorBackTrigger, pinSensorBackEcho)
	// todo check motor pins
	c.motors.left = gpio.NewMotor(pinMotorLeftPlus, pinMotorLeftMinus, pinMotorLeftEnable)
	c.motors.right = gpio.NewMotor(pinMotorRightPlus, pinMotorRightMinus, pinMotorRightEnable)

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
	c.motors.left.Clockwize()
	c.motors.right.Clockwize()

	c.lights.front.Blink()
	glog.Info("Moving forwards")
}

func (c *car) backwards() {
	c.motors.left.CounterClockwize()
	c.motors.right.CounterClockwize()

	c.lights.back.Blink()
	glog.Info("Moving backwards")
}

func (c *car) turnLeft() {
	c.motors.left.CounterClockwize()
	c.motors.right.Clockwize()

	c.lights.front.Blink()
	glog.Info("Left turn")
}

func (c *car) turnRight() {
	c.motors.left.Clockwize()
	c.motors.right.CounterClockwize()

	c.lights.front.Blink()
	glog.Info("Right turn")
}

func (c *car) stop() {
	c.motors.left.Stop()
	c.motors.right.Stop()
	glog.Info("Stopped moving")

}
