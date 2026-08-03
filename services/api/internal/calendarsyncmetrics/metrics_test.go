package calendarsyncmetrics

import "testing"

func TestObserveEventDecisionRoutesReasons(t *testing.T) {
	r := New()
	r.ObserveEventDecision(true, ReasonETag)
	r.ObserveEventDecision(true, ReasonProviderUpdatedAt)
	r.ObserveEventDecision(true, ReasonContentHash)
	r.ObserveEventDecision(false, ReasonETagChanged)
	r.ObserveEventDecision(false, ReasonContentHash)

	snap := r.Snapshot()
	if snap.EventSkipETagTotal != 1 || snap.EventSkipUpdatedAtTotal != 1 || snap.EventSkipHashTotal != 1 {
		t.Fatalf("skip metrics = %+v", snap)
	}
	if snap.EventUpdateTotal != 2 {
		t.Fatalf("update total = %d", snap.EventUpdateTotal)
	}
}
