package processor

import (
	"context"
	"github.com/stretchr/testify/assert"
	"log/slog"
	"testing"
)

func TestProcessor_Health(t *testing.T) {
	p := New(nil, slog.Default())
	ctx := context.Background()

	tt := []struct {
		name       string
		preActions func(p *Processor)
		want       Health
	}{
		{
			name:       "empty",
			preActions: func(p *Processor) {},
			want:       Health{},
		},
		{
			name: "adding",
			preActions: func(p *Processor) {
				p.received.Add(5)
				p.accepted.Add(2)
				p.processing.Add("foo")
				p.processing.Add("bar")
			},
			want: Health{Received: 5, Accepted: 2, Processing: []string{"bar", "foo"}},
		},
		{
			name: "removing",
			preActions: func(p *Processor) {
				p.processing.Remove("foo")
				p.processing.Remove("bar")
			},
			want: Health{Received: 5, Accepted: 2},
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			tc.preActions(p)
			assert.Equal(t, tc.want, p.Health(ctx))
		})
	}
}
