package testing

import (
	"fmt"
	"math/rand"
)

const (
	letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	numBytes    = "0123456789"
)

func RandomUint() uint {
	return uint(rand.Intn(1000000))
}

// testPkg.RandomName returns a random string of 10 characters, useful for generating user names in tests.
func RandomName() string {
	return RandomString(10)
}

// testPkg.RandomEmail returns a random email address, useful for generating email addresses in tests.
func RandomEmail() string {
	name := RandomString(8)
	domain := "example.com"
	return fmt.Sprintf("%s@%s", name, domain)
}

// RandomPassword returns a random password of a length between minLength and maxLength.
// The password is composed of a mix of letters and numbers.
// The complexity of the password can be adjusted by changing the values of minLength and maxLength.
func RandomPassword() string {
	// Adjust these parameters as needed for your password complexity requirements
	minLength := 12
	maxLength := 16

	length := rand.Intn(maxLength-minLength) + minLength

	charPool := []rune{}
	charPool = append(charPool, []rune(letterBytes)...)
	charPool = append(charPool, []rune(numBytes)...)

	password := ""
	for i := 0; i < length; i++ {
		password += string(charPool[rand.Intn(len(charPool))])
	}

	return password
}

// testPkg.RandomString returns a random string of n characters, consisting only of letters.
func RandomString(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}
