package scheduler

import (
	"context"
	"testing"
)

// A deck of 10,000 cards (7,000 studied, 3,000 new) over 365 days, at one
// retention and for Help me decide's 30.
func BenchmarkSimulate10k(b *testing.B) {
	deck := simDeck(7000, 3000)
	for b.Loop() {
		if _, err := Simulate(context.Background(), deck, simConfig(0.9)); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkSuggestRetention10k(b *testing.B) {
	deck := simDeck(7000, 3000)
	for b.Loop() {
		if _, err := SuggestRetention(context.Background(), deck, simConfig(0), func(int, int) {}); err != nil {
			b.Fatal(err)
		}
	}
}
