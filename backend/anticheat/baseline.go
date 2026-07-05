package anticheat

import (
	"chemistryuno/backend/database"
	"log"
	"math"
	"time"
)

// BaselineSample carries the per-game indicator values used to update a player's
// behavioral baseline. Only non-violation games should be passed in.
type BaselineSample struct {
	PlayerUID uint
	// ResponseTimes are the raw operation response times (ms) from the game.
	ResponseTimes []int64
	// WinRate is the player's win rate observed up to and including this game (0-1).
	WinRate float64
	// HasWinRate indicates whether WinRate is meaningful for this sample.
	HasWinRate bool
	// IsViolation marks the sample as part of a confirmed-violation period; such
	// samples are excluded from the baseline.
	IsViolation bool
	SampledAt   time.Time
}

// BaselineCollector maintains rolling per-player behavioral baselines. It is
// gated by the adaptive-threshold feature toggle; when disabled it is a no-op.
type BaselineCollector struct {
	repo   baselineStore
	config func() AdaptiveThresholdConfig
}

// baselineStore is the persistence surface the collector needs. The concrete
// *repository.CheatRepository satisfies it.
type baselineStore interface {
	GetPlayerBaseline(playerUID uint, indicator string) (*database.PlayerBehaviorBaseline, error)
	UpsertPlayerBaseline(baseline *database.PlayerBehaviorBaseline) error
}

// NewBaselineCollector creates a collector. configFn returns the live adaptive
// threshold config so window/enable changes take effect without restart.
func NewBaselineCollector(repo baselineStore, configFn func() AdaptiveThresholdConfig) *BaselineCollector {
	return &BaselineCollector{repo: repo, config: configFn}
}

// indicatorResponseTimeMean and friends name the baseline indicators tracked.
const (
	baselineIndicatorResponseMean = "response_time_mean"
	baselineIndicatorWinRate      = "win_rate"
)

// Collect updates the player's baselines from a single game sample. It returns
// early (no writes) when the feature is disabled or when the sample is a
// confirmed-violation sample.
func (bc *BaselineCollector) Collect(sample BaselineSample) error {
	if bc == nil || bc.repo == nil {
		return nil
	}
	cfg := bc.config()
	if !cfg.Enabled {
		return nil
	}
	// Confirmed-violation samples must never pollute the baseline.
	if sample.IsViolation {
		return nil
	}
	if sample.PlayerUID == 0 {
		return nil
	}
	when := sample.SampledAt
	if when.IsZero() {
		when = time.Now()
	}

	if len(sample.ResponseTimes) > 0 {
		mean := meanInt64(sample.ResponseTimes)
		if err := bc.updateIndicator(sample.PlayerUID, baselineIndicatorResponseMean, mean, cfg, when); err != nil {
			log.Printf("[基线采集] 更新 %s 失败: %v", baselineIndicatorResponseMean, err)
		}
	}
	if sample.HasWinRate {
		if err := bc.updateIndicator(sample.PlayerUID, baselineIndicatorWinRate, sample.WinRate, cfg, when); err != nil {
			log.Printf("[基线采集] 更新 %s 失败: %v", baselineIndicatorWinRate, err)
		}
	}
	return nil
}

// updateIndicator applies a rolling mean/variance update bounded by the
// configured window. The window caps the effective sample count so the baseline
// tracks recent behavior rather than an unbounded lifetime average.
func (bc *BaselineCollector) updateIndicator(playerUID uint, indicator string, value float64, cfg AdaptiveThresholdConfig, when time.Time) error {
	existing, err := bc.repo.GetPlayerBaseline(playerUID, indicator)
	if err != nil {
		return err
	}

	windowKind := cfg.BaselineWindowKind
	if windowKind == "" {
		windowKind = "count"
	}

	if existing == nil {
		return bc.repo.UpsertPlayerBaseline(&database.PlayerBehaviorBaseline{
			PlayerUID:     playerUID,
			Indicator:     indicator,
			Mean:          value,
			Variance:      0,
			SampleCount:   1,
			WindowSize:    cfg.BaselineWindow,
			WindowKind:    windowKind,
			LastSampledAt: when,
		})
	}

	// Cap the effective N at the window so the baseline stays responsive.
	n := existing.SampleCount
	if cfg.BaselineWindow > 0 && n >= cfg.BaselineWindow {
		n = cfg.BaselineWindow - 1
		if n < 1 {
			n = 1
		}
	}

	newMean, newVar := rollingMeanVariance(existing.Mean, existing.Variance, n, value)
	newCount := existing.SampleCount + 1
	if cfg.BaselineWindow > 0 && newCount > cfg.BaselineWindow {
		newCount = cfg.BaselineWindow
	}

	return bc.repo.UpsertPlayerBaseline(&database.PlayerBehaviorBaseline{
		PlayerUID:     playerUID,
		Indicator:     indicator,
		Mean:          newMean,
		Variance:      newVar,
		SampleCount:   newCount,
		WindowSize:    cfg.BaselineWindow,
		WindowKind:    windowKind,
		LastSampledAt: when,
	})
}

// rollingMeanVariance computes an incremental mean and population variance using
// Welford-style updates against a prior aggregate of n samples.
func rollingMeanVariance(prevMean, prevVar float64, n int, value float64) (float64, float64) {
	if n <= 0 {
		return value, 0
	}
	count := float64(n)
	newMean := prevMean + (value-prevMean)/(count+1)
	// prevVar is population variance over n samples; reconstruct M2 and extend.
	prevM2 := prevVar * count
	newM2 := prevM2 + (value-prevMean)*(value-newMean)
	newVar := newM2 / (count + 1)
	if newVar < 0 {
		newVar = 0
	}
	return newMean, newVar
}

func meanInt64(values []int64) float64 {
	if len(values) == 0 {
		return 0
	}
	var total int64
	for _, v := range values {
		total += v
	}
	return float64(total) / float64(len(values))
}

// StdDev returns the standard deviation implied by a baseline's variance.
func baselineStdDev(variance float64) float64 {
	if variance <= 0 {
		return 0
	}
	return math.Sqrt(variance)
}
