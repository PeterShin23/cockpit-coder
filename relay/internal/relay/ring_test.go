package relay

import (
	"testing"
)

func TestRingPushAndReplay(t *testing.T) {
	r := NewRing(100)

	data1 := []byte(`{"json":1}`)
	data2 := []byte(`{"json":2}`)
	data3 := []byte(`{"json":3}`)

	r.Push(1, data1)
	r.Push(2, data2)
	r.Push(3, data3)

	replay := r.ReplayFrom(0)
	if len(replay) != 3 {
		t.Errorf("Expected 3 frames, got %d", len(replay))
	}
	if string(replay[0]) != string(data1) {
		t.Errorf("Expected data1, got %s", replay[0])
	}

	replay2 := r.ReplayFrom(2)
	if len(replay2) != 1 {
		t.Errorf("Expected 1 frame, got %d", len(replay2))
	}
	if string(replay2[0]) != string(data3) {
		t.Errorf("Expected data3, got %s", replay2[0])
	}

	// Overflow
	largeData := make([]byte, 150)
	r.Push(4, largeData)
	replay3 := r.ReplayFrom(0)
	if len(replay3) < 1 {
		t.Errorf("Expected some frames, got %d", len(replay3))
	}
}

func TestRingCapacity(t *testing.T) {
	r := NewRing(10)

	smallData := []byte("a")
	for i := 0; i < 20; i++ {
		r.Push(uint64(i), smallData)
	}

	replay := r.ReplayFrom(0)
	if len(replay) > 10 {
		t.Errorf("Expected max 10 frames, got %d", len(replay))
	}
}
