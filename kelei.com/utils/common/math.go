package common

import (
	"fmt"
	"math/rand"
	"strconv"
	"time"
)

func Clampf(value, min_inclusive, max_inclusive int) int {
	temp := 0
	if min_inclusive > max_inclusive {
		temp = min_inclusive
		min_inclusive = max_inclusive
		max_inclusive = temp
	}

	if value < min_inclusive {
		return min_inclusive
	} else if value < max_inclusive {
		return value
	} else {
		return max_inclusive
	}
}

func Round2(f float64, n int) float64 {
	floatStr := fmt.Sprintf("%."+strconv.Itoa(n)+"f", f)
	inst, _ := strconv.ParseFloat(floatStr, 64)
	return inst
}

func Perm(n int) []int {
	rr := rand.New(rand.NewSource(time.Now().UnixNano()))
	indexs := rr.Perm(n)
	return indexs
}
