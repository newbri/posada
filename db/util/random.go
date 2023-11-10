package util

import (
	"fmt"
	"math/rand"
	"strings"
	"time"
)

var r *rand.Rand

const alphabet = "abcdefghijklmnopqrstuvzxyz"

func init() {
	source := rand.NewSource(time.Now().UnixNano())
	r = rand.New(source)
}

func RandomInt(min, max int64) int64 {
	return min + r.Int63n(max-min+1)
}

func RandomString(n int) string {
	var sb strings.Builder
	k := len(alphabet)

	for i := 0; i < n; i++ {
		c := alphabet[rand.Intn(k)]
		sb.WriteByte(c)
	}

	return sb.String()
}

func RandomOwner() string {
	return RandomString(6)
}

func RandomEmail() string {
	return fmt.Sprintf("%s@email.com", RandomString(6))
}
