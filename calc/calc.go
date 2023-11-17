package calc

import (
	"math"
	"time"
)

type Coordinates = struct {
	X float64
	Y float64
}

type Position = struct {
	X float64
	Y float64
	Rotation float64
}

type Camera = struct {
	Position
	Fov float64
	LastX float64
}

type TimedPosition = struct {
	Coordinates

	Time time.Time
}

func Extrapolate(a, b TimedPosition) TimedPosition {
	now := time.Now()

	td := now.Sub(a.Time)
	tmax := b.Time.Sub(a.Time)
	t := td.Seconds() / tmax.Seconds()
	
	x, y := lerp(a.Coordinates, b.Coordinates, t)
	return TimedPosition{ Coordinates: Coordinates { X: x, Y: y}, Time: now } 
}

func lerp(a, b Coordinates, t float64) (float64, float64) {
	return a.X + (b.X - a.X) * t, a.Y + (b.Y - a.Y) * t
}

func rad(x float64) float64 {
	return x / 180 * math.Pi
}

func Calc(ax, ay, bx, by, x1, x2, fova, fovb, rota, rotb float64) (bool, float64, float64) {
	alpha := rad(rota + (fova * (0.5 - x1)))
	beta := rad(rotb + (fovb * (0.5 - x2)))
	
	tana := math.Tan(alpha)
	tanb := math.Tan(beta)

	return intersect(tana, tanb, ax, ay, bx, by)
}

func intersect(aslope, bslope, ax, ay, bx, by float64) (bool, float64, float64) {
	x := (aslope * ax - bslope * bx + by - ay) / (aslope - bslope)
	y := aslope * (x - ax) + ay
	
	// util.D("x: %v, y: %v\n", x, y)
	
	if math.IsNaN(x) || math.IsInf(x, 0) || math.IsNaN(y) || math.IsInf(y, 0) {
		return false, 0, 0
	}
	
	return true, x, y
}

func CheckSetup(a, b Camera) bool {
	acenter := math.Tan(rad(a.Rotation))
	bcenter := math.Tan(rad(b.Rotation))

	ok, x, y := intersect(acenter, bcenter, a.X, a.Y, b.X, b.Y)
	if !ok {
		return false
	}
	
	q1 := getQuadrant(a.Rotation)
	q2 := getQuadrant(b.Rotation)

	if (q1[0] > 0 && !(x >= a.X)) || (q1[0] < 0 && !(x <= a.X)) {
		return false
	}
	if (q1[1] > 0 && !(y >= a.Y)) || (q1[1] < 0 && !(y <= a.Y)) {
		return false
	}

	if (q2[0] > 0 && !(x >= b.X)) || (q2[0] < 0 && !(x <= b.X)) {
		return false
	}
	if (q2[1] > 0 && !(y >= b.Y)) || (q2[1] < 0 && !(y <= b.Y)) {
		return false
	}

	return true
}

func getQuadrant(angle float64) [2]float64 {
	quadrants := [][2]float64{ { 1, 1 }, { -1, 1 }, { -1, -1 }, { 1, -1 } }

	if angle >= 0 {
		for i := 0; i < 4; i++ {
			if angle <= float64((i + 1) * 90) {
				return quadrants[i]
			}
		}
	} else {
		for i := 0; i < 4; i++ {
			if angle >= -float64((i + 1) * 90) {
				return quadrants[3 - i]
			}
		}
	}

	return [2]float64{ 0, 0 }
}
