package yield

import (
	"testing"
	"time"
)

func TestEveryCadence(t *testing.T) {
	calls := 0
	pause = func() { calls++ }
	defer func() { pause = Yield }()
	f := Every(4)
	for i := 0; i < 10; i++ {
		f()
	}
	if calls != 2 {
		t.Fatalf("yielded %d times in 10 calls, want 2", calls)
	}
}

func TestYieldIsCheapNatively(t *testing.T) {
	start := time.Now()
	for i := 0; i < 1000; i++ {
		Yield()
	}
	if time.Since(start) > 50*time.Millisecond && !isJS {
		t.Fatal("Yield blocks on native")
	}
}
