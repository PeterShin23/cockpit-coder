package relay

import "sync"

type Frame struct {
	Seq  uint64
	Data []byte
}

type Ring struct {
	mu       sync.Mutex
	capacity int
	total    int
	frames   []Frame
}

func NewRing(capacity int) *Ring {
	return &Ring{
		capacity: capacity,
		frames:   make([]Frame, 0),
		total:    0,
	}
}

func (r *Ring) Push(seq uint64, data []byte) {
	r.mu.Lock()
	defer r.mu.Unlock()

	for len(r.frames) > 0 && r.total+len(data) > r.capacity {
		old := r.frames[0]
		r.total -= len(old.Data)
		r.frames = r.frames[1:]
	}

	r.frames = append(r.frames, Frame{
		Seq:  seq,
		Data: append([]byte(nil), data...), // copy to own
	})
	r.total += len(data)
}

func (r *Ring) ReplayFrom(seq uint64) [][]byte {
	r.mu.Lock()
	defer r.mu.Unlock()

	var res [][]byte
	for _, f := range r.frames {
		if f.Seq > seq {
			res = append(res, append([]byte(nil), f.Data...)) // copy
		}
	}
	return res
}
