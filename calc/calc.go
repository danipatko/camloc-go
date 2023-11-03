package calc

import (
	"math"
	"time"
)

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
	Position

	Time time.Time
}

func Extrapolate(a, b TimedPosition) (float64, float64) {
	now := time.Now()

	td := now.Sub(a.Time)
	tmax := b.Time.Sub(a.Time)
	t := td.Seconds() / tmax.Seconds()
	
	x, y := lerp(a.Position, b.Position, t)
	return x, y
}

func lerp(a, b Position, t float64) (float64, float64) {
	return a.X + (b.X - a.X) * t, a.Y + (b.Y - a.Y) * t
}

const MIN_DIFF = 0.

func CalcPosition(data map[string]Camera) *Position {
	if len(data) < 2 {
		return nil
	}

	tangents := map[string]float64{}
	for k, v := range data {
		tangents[k] = math.Tan(v.Position.Rotation + (v.Fov * (0.5 - v.LastX)))
	}

	x, y, points := 0., 0., 0.

	idx := 0
	for i, v := range data {
		atan := tangents[i]
		c1 := v.Position
		a1 := math.Mod(c1.Rotation, 180.)

		jdx := 0
		for j, t := range data {
			if jdx > idx {
				continue
			}
			jdx++

			btan := tangents[j]
			c2 := t.Position
			a2 := math.Mod(c2.Rotation, 180.)

			diff := math.Abs(a1 - a2)
			if diff < MIN_DIFF {
				continue
			}
			
			px := (c1.X * atan - c2.X * btan - c1.Y + c2.Y) / (atan - btan)
			py := atan * (px - c1.X) + c1.Y
			
			x += px
			y += py
			
			points++
		}

		// for j := 0; j < i; j++ {
		// 	btan := tangents[j]
		// 	c2 := data[j].Position
		// 	a2 := math.Mod(c2.Rotation, 180.)

		// 	diff := math.Abs(a1 - a2)
		// 	if diff < MIN_DIFF {
		// 		continue
		// 	}

		// 	px := (c1.X * atan - c2.X * btan - c1.Y + c2.Y) / (atan - btan)
		// 	py := atan * (px - c1.X) + c1.Y

		// 	x += px
		// 	y += py

		// 	points++
		// }

		idx++
	}

	x /= points
	y /= points

	return &Position{ X: x, Y: y }
}


