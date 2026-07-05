package anticheat

import (
	"chemistryuno/backend/database"
	"fmt"
	"testing"
)

// fakeBaselineStore is an in-memory baselineStore for unit tests.
type fakeBaselineStore struct {
	rows map[string]*database.PlayerBehaviorBaseline
}

func newFakeBaselineStore() *fakeBaselineStore {
	return &fakeBaselineStore{rows: make(map[string]*database.PlayerBehaviorBaseline)}
}

func (f *fakeBaselineStore) key(uid uint, indicator string) string {
	return fmt.Sprintf("%d:%s", uid, indicator)
}

func (f *fakeBaselineStore) GetPlayerBaseline(uid uint, indicator string) (*database.PlayerBehaviorBaseline, error) {
	if row, ok := f.rows[f.key(uid, indicator)]; ok {
		clone := *row
		return &clone, nil
	}
	return nil, nil
}

func (f *fakeBaselineStore) UpsertPlayerBaseline(b *database.PlayerBehaviorBaseline) error {
	clone := *b
	f.rows[f.key(b.PlayerUID, b.Indicator)] = &clone
	return nil
}

func adaptiveCfg(enabled bool, window int) func() AdaptiveThresholdConfig {
	return func() AdaptiveThresholdConfig {
		return AdaptiveThresholdConfig{
			Enabled:            enabled,
			BaselineWindow:     window,
			BaselineWindowKind: "count",
			MinSamples:         5,
			PersonalWeight:     0.5,
			GlobalSuperhumanZ:  3.0,
		}
	}
}

func TestBaselineCollector_DisabledIsNoop(t *testing.T) {
	store := newFakeBaselineStore()
	bc := NewBaselineCollector(store, adaptiveCfg(false, 50))
	if err := bc.Collect(BaselineSample{PlayerUID: 1, ResponseTimes: []int64{100, 120}}); err != nil {
		t.Fatalf("collect: %v", err)
	}
	if got, _ := store.GetPlayerBaseline(1, baselineIndicatorResponseMean); got != nil {
		t.Fatalf("expected no baseline when feature disabled, got %+v", got)
	}
}

func TestBaselineCollector_UpdatesMean(t *testing.T) {
	store := newFakeBaselineStore()
	bc := NewBaselineCollector(store, adaptiveCfg(true, 50))

	// First sample: mean = 100.
	if err := bc.Collect(BaselineSample{PlayerUID: 1, ResponseTimes: []int64{100}}); err != nil {
		t.Fatalf("collect: %v", err)
	}
	row, _ := store.GetPlayerBaseline(1, baselineIndicatorResponseMean)
	if row == nil || row.Mean != 100 || row.SampleCount != 1 {
		t.Fatalf("after first sample expected mean=100 count=1, got %+v", row)
	}

	// Second sample of 200 -> rolling mean should move toward 150.
	if err := bc.Collect(BaselineSample{PlayerUID: 1, ResponseTimes: []int64{200}}); err != nil {
		t.Fatalf("collect: %v", err)
	}
	row, _ = store.GetPlayerBaseline(1, baselineIndicatorResponseMean)
	if row.SampleCount != 2 {
		t.Fatalf("expected count=2, got %d", row.SampleCount)
	}
	if row.Mean < 140 || row.Mean > 160 {
		t.Fatalf("expected rolling mean near 150, got %.2f", row.Mean)
	}
	if row.Variance <= 0 {
		t.Fatalf("expected positive variance after differing samples, got %.4f", row.Variance)
	}
}

func TestBaselineCollector_ExcludesViolationSamples(t *testing.T) {
	store := newFakeBaselineStore()
	bc := NewBaselineCollector(store, adaptiveCfg(true, 50))

	if err := bc.Collect(BaselineSample{PlayerUID: 7, ResponseTimes: []int64{50}, IsViolation: true}); err != nil {
		t.Fatalf("collect: %v", err)
	}
	if got, _ := store.GetPlayerBaseline(7, baselineIndicatorResponseMean); got != nil {
		t.Fatalf("violation sample must not update baseline, got %+v", got)
	}
}

func TestBaselineCollector_WindowCapsSampleCount(t *testing.T) {
	store := newFakeBaselineStore()
	window := 5
	bc := NewBaselineCollector(store, adaptiveCfg(true, window))

	for i := 0; i < 20; i++ {
		if err := bc.Collect(BaselineSample{PlayerUID: 3, ResponseTimes: []int64{100}}); err != nil {
			t.Fatalf("collect %d: %v", i, err)
		}
	}
	row, _ := store.GetPlayerBaseline(3, baselineIndicatorResponseMean)
	if row.SampleCount != window {
		t.Fatalf("expected sample count capped at window=%d, got %d", window, row.SampleCount)
	}
}

func TestBaselineCollector_WinRateIndicator(t *testing.T) {
	store := newFakeBaselineStore()
	bc := NewBaselineCollector(store, adaptiveCfg(true, 50))

	if err := bc.Collect(BaselineSample{PlayerUID: 9, WinRate: 0.5, HasWinRate: true}); err != nil {
		t.Fatalf("collect: %v", err)
	}
	row, _ := store.GetPlayerBaseline(9, baselineIndicatorWinRate)
	if row == nil || row.Mean != 0.5 {
		t.Fatalf("expected win_rate baseline mean=0.5, got %+v", row)
	}
}
