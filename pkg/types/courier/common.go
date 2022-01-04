package courier

import "time"

const (
	defaultConnCheckInterval = 30 * time.Second
	defaultReadDeadline      = 60 * time.Second
	defaultWriteDeadline     = 60 * time.Second
	defaultWaitTime          = 1 * time.Second
	maxWaitTime              = 10 * time.Second
)
