package nats

import (
	"testing"
	"time"

	"review-service/config"
)

func TestPullConsumerConfigValidator(t *testing.T) {
	v := NewPullConsumerConfigValidator()

	valid := config.PullConsumerConfig{
		Stream:     "stream",
		Subject:    "subject",
		Durable:    "durable",
		BatchSize:  1,
		MaxWait:    time.Second,
		Workers:    1,
		QueueSize:  1,
		AckWait:    time.Second,
		MaxDeliver: 1,
	}
	if err := v.Validate(valid); err != nil {
		t.Fatalf("expected valid config, got %v", err)
	}

	tests := []struct {
		name string
		cfg  config.PullConsumerConfig
	}{
		{name: "empty stream", cfg: func() config.PullConsumerConfig { c := valid; c.Stream = ""; return c }()},
		{name: "empty subject", cfg: func() config.PullConsumerConfig { c := valid; c.Subject = ""; return c }()},
		{name: "empty durable", cfg: func() config.PullConsumerConfig { c := valid; c.Durable = ""; return c }()},
		{name: "bad batch", cfg: func() config.PullConsumerConfig { c := valid; c.BatchSize = 0; return c }()},
		{name: "bad max wait", cfg: func() config.PullConsumerConfig { c := valid; c.MaxWait = 0; return c }()},
		{name: "bad workers", cfg: func() config.PullConsumerConfig { c := valid; c.Workers = 0; return c }()},
		{name: "bad queue size", cfg: func() config.PullConsumerConfig { c := valid; c.QueueSize = 0; return c }()},
		{name: "bad ack wait", cfg: func() config.PullConsumerConfig { c := valid; c.AckWait = 0; return c }()},
		{name: "bad max deliver", cfg: func() config.PullConsumerConfig { c := valid; c.MaxDeliver = 0; return c }()},
		{name: "bad deliver policy", cfg: func() config.PullConsumerConfig { c := valid; c.DeliverPolicy = "invalid"; return c }()},
		{name: "bad adaptive interval", cfg: func() config.PullConsumerConfig {
			c := valid
			c.Adaptive.Enabled = true
			c.Adaptive.CheckInterval = 0
			return c
		}()},
		{name: "bad adaptive thresholds", cfg: func() config.PullConsumerConfig {
			c := valid
			c.Adaptive.Enabled = true
			c.Adaptive.CheckInterval = time.Second
			c.Adaptive.MediumPending = 3
			c.Adaptive.HighPending = 3
			return c
		}()},
		{name: "bad adaptive plan", cfg: func() config.PullConsumerConfig {
			c := valid
			c.Adaptive.Enabled = true
			c.Adaptive.CheckInterval = time.Second
			c.Adaptive.MediumPending = 1
			c.Adaptive.HighPending = 2
			c.Adaptive.LowBatchSize = 0
			return c
		}()},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if err := v.Validate(tc.cfg); err == nil {
				t.Fatalf("expected error")
			}
		})
	}
}
