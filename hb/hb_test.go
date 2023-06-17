package hb

import (
	"testing"
	"time"
)

func TestSince(t *testing.T) {
	type test struct {
		input HBTime
		now   time.Time
		want  string
	}

	now := time.Now()
	tests := []test{
		{
			input: HBTime(now.Add(time.Hour * 24 * 365 * -6)),
			now:   now,
			want:  "6 years ago",
		},
		{
			input: HBTime(now.Add(time.Hour * -50)),
			now:   now,
			want:  "2 days ago",
		},
		{
			input: HBTime(now.Add(time.Hour * -23)),
			now:   now,
			want:  "23 hours ago",
		},
		{
			input: HBTime(now.Add(time.Hour * -5)),
			now:   now,
			want:  "5 hours ago",
		},
		{
			input: HBTime(now.Add(time.Minute * -5)),
			now:   now,
			want:  "5 minutes ago",
		},
		{
			input: HBTime(now.Add(time.Second * -5)),
			now:   now,
			want:  "5 seconds ago",
		},
		{
			input: HBTime(now),
			now:   now,
			want:  "0 seconds ago",
		},
		{
			input: HBTime(now.Add(time.Millisecond * -5)),
			now:   now,
			want:  "0 seconds ago",
		},
	}

	for _, tc := range tests {
		got := tc.input.Since(tc.now)
		if got != tc.want {
			t.Fatalf("got %s want %s", got, tc.want)
		}
	}
}
