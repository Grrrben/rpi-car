package rpi_car

import (
	"fmt"
	"github.com/golang/glog"
	"github.com/grrrben/gpio"
	"time"
)

const strFront = "F"
const strBack = "B"
const strLeft = "L"
const strRight = "R"

// used pins in BCM numbering
const pinLedFront = 10
const pinLedBack = 9

const pinSensorFrontTrigger = 20
const pinSensorFrontEcho = 21

const pinSensorBackTrigger = 8
const pinSensorBackEcho = 7

const pinSensorLeftTrigger = 19
const pinSensorLeftEcho = 26

const pinSensorRightTrigger = 6
const pinSensorRightEcho = 13

const pinMotorLeftPlus = 17
const pinMotorLeftMinus = 18
const pinMotorLeftEnable = 27

const pinMotorRightPlus = 22
const pinMotorRightMinus = 23
const pinMotorRightEnable = 24

const sensorStable = 3

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
	front *gpio.HCSR04
	back  *gpio.HCSR04
	left  *gpio.HCSR04
	right *gpio.HCSR04
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
	c.sensors.left = gpio.NewHCSR04(pinSensorLeftTrigger, pinSensorLeftEcho)
	c.sensors.right = gpio.NewHCSR04(pinSensorRightTrigger, pinSensorRightEcho)

	// todo check motor pins
	c.motors.left = gpio.NewMotor(pinMotorLeftPlus, pinMotorLeftMinus, pinMotorLeftEnable)
	c.motors.right = gpio.NewMotor(pinMotorRightPlus, pinMotorRightMinus, pinMotorRightEnable)

	return c
}

func (c *car) Init() {
	fmt.Println("Car initiated")

	c.stop()
	c.lights.front.BlinkBlink(3)

	glog.Info("1 sec break")
	time.Sleep(time.Second)

	c.drive()
}

// caching last distances for taking measurements
type lastKnownDistances struct {
	front, back, left, right float64
}

func (c *car) drive() {

	sensorTimeout := time.Millisecond * 200

	front := make(chan float64, 3)
	back := make(chan float64, 3)
	left := make(chan float64, 3)
	right := make(chan float64, 3)

	go func() {
		for {
			front <- c.sensors.front.Measure()
			time.Sleep(sensorTimeout)
			back <- c.sensors.back.Measure()
			time.Sleep(sensorTimeout)
			left <- c.sensors.left.Measure()
			time.Sleep(sensorTimeout)
			right <- c.sensors.right.Measure()
			time.Sleep(sensorTimeout)
		}
	}()

	lkd := lastKnownDistances{
		front: 101,
		back:  101,
		left:  101,
		right: 101,
	}

	for {
		select {
		case lkd.front = <-front:
			glog.Infof("%s: %.2f\n", strFront, lkd.front)
			c.decide(lkd)
		case lkd.back = <-back:
			glog.Infof("%s:          %.2f\n", strBack, lkd.back)
			c.decide(lkd)
		case lkd.left = <-left:
			glog.Infof("%s:                   %.2f\n", strLeft, lkd.left)
			c.decide(lkd)
		case lkd.right = <-right:
			glog.Infof("%s:                            %.2f\n", strRight, lkd.right)
			c.decide(lkd)
		}
	}
}

func (c *car) decide(lkd lastKnownDistances) {
	d, cm := getMinDistance(lkd)

	if cm > 100 {
		c.stop()
		return
	}

	c.moveInOpposideDirection(d)
}

func (c *car) moveInOpposideDirection(d string) {
	switch d {
	case strFront:
		c.backwards()
	case strBack:
		c.forwards()
	case strLeft:
		c.turnRight()
		time.Sleep(turnTime)
	case strRight:
		c.turnLeft()
		time.Sleep(turnTime)
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
	c.motors.right.Clockwize()
	glog.Info("Right")
}

func (c *car) turnLeft() {
	c.motors.right.CounterClockwize()
	c.motors.left.Clockwize()
	glog.Info("Left")
}

func (c *car) stop() {
	c.motors.left.Stop()
	c.motors.right.Stop()
	glog.Info("Stop")

}

func getMinDistance(lkd lastKnownDistances) (d string, cm float64) {
	cm = lkd.front
	d = strFront

	if lkd.back < cm {
		cm = lkd.back
		d = strBack
	}

	if lkd.left < cm {
		cm = lkd.left
		d = strLeft
	}

	if lkd.right < cm {
		cm = lkd.right
		d = strRight
	}
	return
}
