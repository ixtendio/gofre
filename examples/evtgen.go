package main

import (
	"context"
	"fmt"
	"github.com/ixtendio/gofre/response"
	"sync/atomic"
	"time"
)

type EventGenerator struct {
	readIndex  int64
	writeIndex int64
	list       []response.ServerSentEvent
}

func (l *EventGenerator) push(evt response.ServerSentEvent) {
	i := atomic.LoadInt64(&l.writeIndex) % int64(cap(l.list))
	l.list[i] = evt
	atomic.AddInt64(&l.writeIndex, 1)
}

func (l *EventGenerator) Next() (response.ServerSentEvent, bool) {
	ri := atomic.LoadInt64(&l.readIndex)
	wi := atomic.LoadInt64(&l.writeIndex)
	if ri >= wi {
		return response.ServerSentEvent{}, false
	}

	i := ri % int64(cap(l.list))
	evt := l.list[i]
	atomic.AddInt64(&l.readIndex, 1)
	return evt, true
}

func (l *EventGenerator) Rewind(index int) {
	l.readIndex = int64(index)
}

func NewEventGenerator(ctx context.Context, cap int) *EventGenerator {
	evtGen := &EventGenerator{
		list: make([]response.ServerSentEvent, cap, cap),
	}
	go func() {
		var msgIndex int64
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				msgId := fmt.Sprintf("%d", atomic.AddInt64(&msgIndex, 1))
				evtGen.push(response.ServerSentEvent{
					Id:   msgId,
					Name: "message",
					Data: []string{"message " + msgId},
				})
			}
		}

	}()
	return evtGen
}
