package router

import "time"

type ctxKey int

const KeyValues ctxKey = 1

type CtxValues struct {
	CorrelationId string
	StartTime     time.Time
}
