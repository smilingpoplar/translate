package util

import (
	"math/rand"
	"time"
)

var rng *rand.Rand

func init() {
	source := rand.NewSource(time.Now().UnixNano())
	rng = rand.New(source)
}

func RandInt(a, b int) int {
	return a + rng.Intn(b-a)
}
