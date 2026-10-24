//go:build !race

package core

// raceDetector is set when the tests run with -race (see race_on_test.go).
const raceDetector = false
