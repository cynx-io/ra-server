package helper

import (
	"math/rand"
	"time"
)

func GenerateRandomNumber(length int) int {
	if length <= 0 {
		return 0
	}

	minimum := 1
	maximum := 1
	for i := 1; i < length; i++ {
		minimum *= 10
		maximum = maximum*10 + 9
	}

	return minimum + rand.Intn(maximum-minimum+1)
}

func GenerateRandomNumberInRange(min, max, seed int) int {
	if min >= max {
		return min
	}

	var r *rand.Rand
	if seed > 0 {
		r = rand.New(rand.NewSource(int64(seed)))
	} else {
		r = rand.New(rand.NewSource(time.Now().UnixNano()))
	}

	return min + r.Intn(max-min+1)
}
