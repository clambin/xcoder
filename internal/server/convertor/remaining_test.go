package convertor

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func Test_calculateRemaining(t *testing.T) {
	tests := []struct {
		name    string
		start   time.Time
		now     time.Time
		elapsed time.Duration
		total   time.Duration
		want    time.Duration
	}{
		{
			name:    "start",
			start:   time.Date(2023, time.December, 6, 0, 0, 0, 0, time.UTC),
			now:     time.Date(2023, time.December, 6, 0, 1, 0, 0, time.UTC),
			elapsed: time.Minute,
			total:   time.Hour,
			want:    59 * time.Minute,
		},
		{
			name:    "quarter",
			start:   time.Date(2023, time.December, 6, 0, 0, 0, 0, time.UTC),
			now:     time.Date(2023, time.December, 6, 0, 1, 0, 0, time.UTC),
			elapsed: 15 * time.Minute,
			total:   time.Hour,
			want:    3 * time.Minute,
		},
		{
			name:    "mid",
			start:   time.Date(2023, time.December, 6, 0, 0, 0, 0, time.UTC),
			now:     time.Date(2023, time.December, 6, 0, 30, 0, 0, time.UTC),
			elapsed: 30 * time.Minute,
			total:   time.Hour,
			want:    30 * time.Minute,
		},
		{
			name:    "done",
			start:   time.Date(2023, time.December, 6, 0, 0, 0, 0, time.UTC),
			now:     time.Date(2023, time.December, 6, 1, 0, 0, 0, time.UTC),
			elapsed: time.Hour,
			total:   time.Hour,
			want:    0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, calculateRemaining(tt.start, tt.now, tt.elapsed, tt.total))
		})
	}

}
