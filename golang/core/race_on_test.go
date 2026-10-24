//go:build race

package core

// raceDetector is set when the tests run with -race, which makes code 5 to
// 10 times slower: the performance tests then skip their time limits.
const raceDetector = true
