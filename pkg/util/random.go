package util

import "math/rand"

// RandomString produces a random string of lower case letters.
func RandomString(num int) string {
	bytes := make([]byte, num)
	for i := 0; i < num; i++ {
		bytes[i] = byte(randomInt(97, 122)) // lowercase letters.
	}
	return string(bytes)
}

func randomInt(min int, max int) int {
	return min + rand.Intn(max-min)
}
