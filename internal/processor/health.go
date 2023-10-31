package processor

import "context"

type Health struct {
	Received   int64
	Accepted   int64
	Processing []string
}

func (p *Processor) Health(_ context.Context) any {
	return Health{
		Received:   p.received.Load(),
		Accepted:   p.accepted.Load(),
		Processing: p.processing.ListOrdered()}
}
