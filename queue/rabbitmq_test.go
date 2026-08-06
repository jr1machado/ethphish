package queue

import "testing"

func TestNextRetryTarget(t *testing.T) {
	cases := []struct {
		retryCount  int
		wantQueue   string
		wantNextInc int
	}{
		{0, MailRetryQueue, 1},
		{MaxProcessingRetries - 1, MailRetryQueue, MaxProcessingRetries},
		{MaxProcessingRetries, MailDeadQueue, MaxProcessingRetries},
		{MaxProcessingRetries + 3, MailDeadQueue, MaxProcessingRetries + 3},
	}
	for _, c := range cases {
		queue, next := nextRetryTarget(c.retryCount)
		if queue != c.wantQueue {
			t.Errorf("nextRetryTarget(%d) queue = %q, want %q", c.retryCount, queue, c.wantQueue)
		}
		if next != c.wantNextInc {
			t.Errorf("nextRetryTarget(%d) next = %d, want %d", c.retryCount, next, c.wantNextInc)
		}
	}
}

func TestRetryCountFromHeaders(t *testing.T) {
	cases := []struct {
		name    string
		headers map[string]interface{}
		want    int
	}{
		{"missing header defaults to zero", nil, 0},
		{"int32 as published by this package", map[string]interface{}{RetryCountHeader: int32(2)}, 2},
		{"int64 as some brokers report it", map[string]interface{}{RetryCountHeader: int64(4)}, 4},
	}
	for _, c := range cases {
		got := retryCountFromHeaders(c.headers)
		if got != c.want {
			t.Errorf("%s: retryCountFromHeaders() = %d, want %d", c.name, got, c.want)
		}
	}
}
