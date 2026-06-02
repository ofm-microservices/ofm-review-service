package pullplan

import (
	"testing"
	"time"

	"review-service/config"
)

func TestResolverResolve(t *testing.T) {
	r := New()
	base := config.PullConsumerConfig{BatchSize: 2, MaxWait: time.Second}
	tier, batch, wait := r.Resolve(base, 0)
	if tier != "base" || batch != 2 || wait != time.Second {
		t.Fatalf("unexpected base resolve: %s %d %s", tier, batch, wait)
	}

	base.Adaptive.Enabled = true
	base.Adaptive.LowBatchSize = 1
	base.Adaptive.LowMaxWait = time.Millisecond
	base.Adaptive.MediumPending = 3
	base.Adaptive.MediumBatchSize = 4
	base.Adaptive.MediumMaxWait = 2 * time.Millisecond
	base.Adaptive.HighPending = 5
	base.Adaptive.HighBatchSize = 6
	base.Adaptive.HighMaxWait = 3 * time.Millisecond

	tier, batch, wait = r.Resolve(base, 6)
	if tier != "high" || batch != 6 || wait != 3*time.Millisecond {
		t.Fatalf("unexpected high resolve: %s %d %s", tier, batch, wait)
	}
	tier, batch, wait = r.Resolve(base, 4)
	if tier != "medium" || batch != 4 || wait != 2*time.Millisecond {
		t.Fatalf("unexpected medium resolve: %s %d %s", tier, batch, wait)
	}
	tier, batch, wait = r.Resolve(base, 1)
	if tier != "low" || batch != 1 || wait != time.Millisecond {
		t.Fatalf("unexpected low resolve: %s %d %s", tier, batch, wait)
	}
}
