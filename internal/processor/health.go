package processor

import "context"

type Health struct {
	Processing []string
	Received   int64
	Accepted   int64
}

func (p *Processor) Health(_ context.Context) any {
	return Health{
		Received:   p.received.Load(),
		Accepted:   p.accepted.Load(),
		Processing: p.processing.ListOrdered()}
}
