package main

// delta returns a value that is step away from cur, without crossing
// min or max or under/overflowing.
func delta(cur, min, max, step int32) int32 {
	if step > 0 && max-cur < step {
		return max
	}
	if step < 0 && cur-min < -step {
		return min
	}
	return cur + step
}

// delta returns a value approximately 5% of the range in the given
// direction, without crossing min or max or under/overflowing.
//
// When closer than two steps from the end of the range, the increment
// is 0.5%.
func delta5(cur, min, max int32, increase bool) int32 {
	step := (max - min) / 20
	if step < 1 {
		step = 1
	}
	if max-cur <= 2*step || cur-min <= 2*step {
		step = (max - min) / 200
		if step < 1 {
			step = 1
		}
	}
	if !increase {
		step = -step
	}
	return delta(cur, min, max, step)
}
