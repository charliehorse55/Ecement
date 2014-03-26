package libcement

import (
	"math"
)


//assumes Finverse is monotonically increasing
func FinverseToF(in []float32) []float32 {
	out := make([]float32, len(in))
	for i := range out {
		target := (float64(i*i)/float64((len(out) - 1)*(len(out) - 1)))
		minDiff := math.MaxFloat64
		lowIndex := 0
		highIndex := 0
		for j := range in {
			diff := math.Abs(target - float64(in[j]))
			if diff < minDiff {
				minDiff = diff
				belowDiff := math.MaxFloat64
				aboveDiff := math.MaxFloat64
				if j > 0 {
					belowDiff = math.Abs(target - float64(in[j - 1]))
				}
				if j < (len(in) - 1) {
					aboveDiff = math.Abs(target - float64(in[j + 1]))
				}
				
				if belowDiff < aboveDiff {
					lowIndex = j - 1
					highIndex = j
				} else {
					lowIndex = j
					highIndex = j + 1
				}
			}
		}
		// rise over run simplifies to 
		slope := (float64(in[highIndex])-float64(in[lowIndex]))*float64(len(in))
		inter := ((target - float64(in[lowIndex]))/slope) + float64(lowIndex)
		//don't interpolate below 0
		if inter < 0 {
			inter = 0
		}
		out[i] = float32(inter/float64(len(out) - 1))
	}
	return out
}

func createSquaredFinverse(size int) []float32 {
	result := make([]float32, size)
	for i := 0; i < size; i++ {
		result[i] = float32(float64(i*i)/float64((size-1)*(size-1)))
	}
	return result
}
