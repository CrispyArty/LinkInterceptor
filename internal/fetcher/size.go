package fetcher

import (
	"fmt"
	"math"
)

type Size struct {
	Bytes int64
}

func (s Size) Humanize() string {
	return HumanizeSize(s.Bytes)
}

func HumanizeSize(size int64) string {
	const unit = 1024
	// fmt.Printf("%v \n", s.Bytes)

	if size < unit {
		return fmt.Sprintf("%d B", size)
	}

	exp := 0
	sizes := "kMGTPE"

	for n := size; n >= unit && exp < len(sizes)-1; n /= unit {
		exp++
	}

	// size := math.Floor(float64(size)/math.Pow(unit, float64(exp))*100) / 100
	// str := strconv.FormatFloat(size, 'f', -1, 64)

	// return fmt.Sprintf("%v %cB", str, sizes[exp-1])
	return fmt.Sprintf("%.2f %cB", float64(size)/math.Pow(unit, float64(exp)), sizes[exp-1])

}
