package rpi_car

import (
	"fmt"
	"github.com/golang/glog"
	"github.com/grrrben/gpio"
	"time"
)

// used pins in BCM numbering
const pinLedFront = 10
const pinLedBack = 9

const pinSensorFrontTrigger = 8
const pinSensorFrontEcho = 7

const pinSensorLeftTrigger = 20
const pinSensorLeftEcho = 21

const pinSensorRightTrigger = 5
const pinSensorRightEcho = 6

const pinMotorLeftPlus = 17
const pinMotorLeftMinus = 18
const pinMotorLeftEnable = 27

const pinMotorRightPlus = 22
const pinMotorRightMinus = 23
const pinMotorRightEnable = 24

const turnTime = time.Millisecond * 100

type lighting struct {
	front *gpio.Led
	back  *gpio.Led
}

type propulsion struct {
	left  *gpio.Motor
	right *gpio.Motor
}

type distanceSensors struct {
	front      *gpio.HCSR04
	left       *gpio.HCSR04
	right      *gpio.HCSR04
	stabilized int
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
	c.sensors.left = gpio.NewHCSR04(pinSensorLeftTrigger, pinSensorLeftEcho)
	c.sensors.right = gpio.NewHCSR04(pinSensorRightTrigger, pinSensorRightEcho)

	// todo check motor pins
	c.motors.left = gpio.NewMotor(pinMotorLeftPlus, pinMotorLeftMinus, pinMotorLeftEnable)
	c.motors.right = gpio.NewMotor(pinMotorRightPlus, pinMotorRightMinus, pinMotorRightEnable)

	c.sensors.stabilized = 0

	return c
}

func (c *car) Init() {
	glog.Info("Car initiated")

	c.stop()
	c.lights.front.BlinkBlink(3)

	glog.Info("1 sec break")
	time.Sleep(time.Second)

	c.drive()
}

func (c *car) drive() {
	front := make(chan float64, 3)
	left := make(chan float64, 3)
	right := make(chan float64, 3)

	go func() {
		for {
			front <- c.sensors.front.Measure()
			time.Sleep(time.Second / 2)
		}
	}()

	go func() {
		for {
			left <- c.sensors.left.Measure()
			time.Sleep(time.Second / 2)
		}
	}()

	go func() {
		for {
			right <- c.sensors.right.Measure()
			time.Sleep(time.Second / 2)
		}
	}()

	// caching last distances for taking measurements
	var lastFront, lastLeft, lastRight float64

	for {
		select {
		case cmFront := <-front:
			lastFront = cmFront
			fmt.Printf("F: %.2f\n", lastFront)
			c.decide(lastFront, lastLeft, lastRight)
		case cmLeft := <-left:
			lastLeft = cmLeft
			fmt.Printf("L: %.2f\n", lastLeft)
		case cmRight := <-right:
			lastRight = cmRight
			fmt.Printf("R: %.2f\n", lastRight)
		}
	}
}

func (c *car) decide(f, l, r float64) {

	if r > 1 && l > 1 && f > 1 && c.sensors.stabilized == 0 {
		c.sensors.stabilized = 3
		fmt.Println("SENSORS STABILIZED")
	} else if c.sensors.stabilized == 0 {
		// waiting to be stabilised
		c.lights.back.Blink()
		return
	}

	if f < 50 || r < 15 || l < 15 {

		if r < 1 && l < 1 && f < 1 {
			c.sensors.stabilized--
		}

		c.stop()

		if l > r {
			c.turnLeft()
		} else {
			c.turnRight()
		}
		// let it turn for a moment and then stop to take a new decision
		time.Sleep(turnTime)
		c.stop()
	} else {
		c.forwards()
	}
}

func (c *car) forwards() {
	c.motors.left.CounterClockwize()
	c.motors.right.CounterClockwize()

	c.lights.front.Blink()
	glog.Info("Forwards")
}

func (c *car) backwards() {
	c.motors.left.Clockwize()
	c.motors.right.Clockwize()

	c.lights.back.Blink()
	glog.Info("Backwards")
}

func (c *car) turnRight() {
	c.motors.left.CounterClockwize()
	glog.Info("Right")
}

func (c *car) turnLeft() {
	c.motors.right.CounterClockwize()
	glog.Info("Left")
}

func (c *car) stop() {
	c.motors.left.Stop()
	c.motors.right.Stop()
	glog.Info("Stop")

}
