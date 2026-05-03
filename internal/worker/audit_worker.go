package worker

import (
	"context"
	"log"
	"time"
)

type AuditEvent struct {
	Action string
	TaskID int
	When   time.Time
}

type AuditWorker struct {
	events <-chan AuditEvent
}

func NewAuditWorker(events <-chan AuditEvent) *AuditWorker {
	return &AuditWorker{events: events}
}

func (w *AuditWorker) Start(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			log.Println("audit worker stopped:", ctx.Err())
			return
		case event := <-w.events:
			log.Printf("audit event => action=%s taskId=%d at=%s\n",
				event.Action, event.TaskID, event.When)
		}
	}
}
