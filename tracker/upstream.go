package tracker

import (
	"time"
)

type Upstream struct {
	Address   string
	Key       string
	KeepAlive time.Time
	Enabled   bool
	Available bool
}

func (u *Upstream) IsAlive(d time.Duration) bool {
	return u.KeepAlive.After(time.Now().Add(-d))
}
