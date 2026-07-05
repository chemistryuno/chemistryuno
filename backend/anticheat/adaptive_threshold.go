package anticheat

import (
	"chemistryuno/backend/database"
	"encoding/json"
	"math"
)

// GlobalBaselineStat holds the population distribution for one indicator, used as
// the "global baseline" track of adaptive thresholding.
type GlobalBaselineStat struct {
	Mean   float64 `json:"mean"`
	StdDev float64 `json:"std_dev"`
}

// AdaptiveDeviation is the dual-track evaluation outcome for one indicator.
type AdaptiveDeviation struct {
	Indicator      string  `json:"indicator"`
	ObservedValue  float64 `json:"observed_value"`
	PersonalMean   float64 `json:"personal_mean"`
	PersonalZ      float64 `json:"personal_z"`      // |deviation| from personal baseline in std units
	GlobalZ        float64 `json:"global_z"`        // |deviation| from global baseline in std units
	ThresholdSource string `json:"threshold_source"` // "personal", "global", or "mixed"
	// Score is the indicator risk contribution in 0-100 derived from the dual tracks.
	Score float64 `json:"score"`
}

// AdaptiveThresholdEvaluator combines a player's personal baseline with the global
// baseline to produce an interpretable, provenance-tagged risk contribution.
type AdaptiveThresholdEvaluator struct {
	personal map[string]database.PlayerBehaviorBaseline
	global   map[string]GlobalBaselineStat
	cfg      AdaptiveThresholdConfig
}

// NewAdaptiveThresholdEvaluator builds an evaluator from a player's baselines and
// the global stats. Either map may be nil.
func NewAdaptiveThresholdEvaluator(personal map[string]database.PlayerBehaviorBaseline, global map[string]GlobalBaselineStat, cfg AdaptiveThresholdConfig) *AdaptiveThresholdEvaluator {
	return &AdaptiveThresholdEvaluator{personal: personal, global: global, cfg: cfg}
}

// Evaluate computes the dual-track deviation for one indicator value.
//
// The personal track measures sudden deviation from the player's own established
// baseline. The global track measures absolute super-human magnitude. A
// consistently abnormal personal baseline cannot suppress the global track: when
// the global z exceeds the configured super-human threshold the contribution is
// raised regardless of the personal comparison.
func (e *AdaptiveThresholdEvaluator) Evaluate(indicator string, observed float64) AdaptiveDeviation {
	out := AdaptiveDeviation{Indicator: indicator, ObservedValue: observed}

	// Personal track (requires a sufficiently established baseline).
	personalZ := 0.0
	personalReady := false
	if base, ok := e.personal[indicator]; ok && base.SampleCount >= e.cfg.MinSamples {
		std := baselineStdDev(base.Variance)
		out.PersonalMean = base.Mean
		if std > 0 {
			personalZ = math.Abs(observed-base.Mean) / std
			personalReady = true
		}
	}
	out.PersonalZ = personalZ

	// Global track.
	globalZ := 0.0
	if g, ok := e.global[indicator]; ok && g.StdDev > 0 {
		globalZ = math.Abs(observed-g.Mean) / g.StdDev
	}
	out.GlobalZ = globalZ

	superhuman := e.cfg.GlobalSuperhumanZ > 0 && globalZ >= e.cfg.GlobalSuperhumanZ

	switch {
	case superhuman:
		// Absolute super-human magnitude dominates and is never suppressed or
		// diluted by the personal baseline. This is the guarantee that a long-term
		// cheater (whose personal baseline is itself abnormal) is still caught.
		out.ThresholdSource = "global"
		out.Score = zToScore(globalZ)
	case personalReady:
		// Below the super-human threshold the global track is treated as normal
		// population variation and contributes nothing on its own; only a sudden
		// deviation from the player's own baseline matters.
		out.ThresholdSource = "personal"
		out.Score = zToScore(personalZ)
	default:
		out.ThresholdSource = ""
		out.Score = 0
	}

	return out
}

// zToScore maps an absolute z-score to a 0-100 risk contribution. A z of 3 maps to
// roughly 75, z of 4+ saturates toward 100.
func zToScore(z float64) float64 {
	if z <= 0 {
		return 0
	}
	score := (z / 4.0) * 100.0
	return clampRiskScore(score)
}

// MarshalBaselineSnapshot serializes the deviations used in a detection for
// provenance, suitable for the CheatRiskScore.BaselineSnapshot column.
func MarshalBaselineSnapshot(devs []AdaptiveDeviation) database.JSON {
	if len(devs) == 0 {
		return nil
	}
	data, _ := json.Marshal(devs)
	return data
}
