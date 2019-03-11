package libsync

import "sync"

type Status interface {
	IsReady() bool
	SetReady()
}

func NewStatus(st bool) Status {
	return &status{status: st}
}

type status struct {
	sync.RWMutex
	status bool
}

func (r *status) IsReady() bool {
	r.RLock()
	defer r.RUnlock()
	return r.status
}

func (r *status) SetReady() {
	r.Lock()
	r.status = true
	r.Unlock()
}
