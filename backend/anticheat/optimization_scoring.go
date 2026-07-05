package anticheat

import (
	"math"
	"time"
)

// applyOptimizationDimensions folds the adaptive-threshold and z-score dimensions
// into the running weighted total. It is a no-op when both features are disabled,
// preserving historical scoring behavior exactly. The caller must hold configLock.
func (rse *RiskScoringEngine) applyOptimizationDimensions(context *DetectionContext, result *RiskScoringResult, weightedTotal, totalWeight float64) (float64, float64) {
	opt := rse.config.Optimization

	// ----- Adaptive threshold dual-track dimension -----
	if opt.AdaptiveThreshold.Enabled {
		eval := NewAdaptiveThresholdEvaluator(context.PersonalBaselines, context.GlobalBaselines, opt.AdaptiveThreshold)
		devs := make([]AdaptiveDeviation, 0, 2)

		// Response-time mean track.
		if len(context.ResponseTimes) > 0 {
			observed := meanInt64(context.ResponseTimes)
			devs = append(devs, eval.Evaluate(baselineIndicatorResponseMean, observed))
		}
		// Win-rate track.
		if context.HasWinRate {
			devs = append(devs, eval.Evaluate(baselineIndicatorWinRate, context.WinRate))
		}

		if len(devs) > 0 {
			// Combine deviation scores into one adaptive dimension (take the max as the
			// strongest evidence; record all for provenance).
			maxScore := 0.0
			source := ""
			for _, d := range devs {
				if d.Score > maxScore {
					maxScore = d.Score
					source = d.ThresholdSource
				}
			}
			result.AdaptiveDeviations = devs
			result.ThresholdSource = source

			weight := opt.AdaptiveThreshold.ContributionWeight
			if weight > 0 && maxScore > 0 {
				if context.AccountAgeDays < 7 {
					weight *= 1.5
				}
				result.Dimensions["adaptive_threshold"] = maxScore
				result.EffectiveWeights["adaptive_threshold"] = weight
				contribution := maxScore * weight
				weightedTotal += contribution
				totalWeight += weight
				result.IndicatorDetails = append(result.IndicatorDetails, RiskIndicatorDetail{
					Name:            "adaptive_threshold",
					RawValue:        maxScore,
					NormalizedScore: clampRiskScore(maxScore),
					Weight:          weight,
					Contribution:    contribution,
					Explanation:     "deviation from personal/global baseline (source: " + source + ")",
					EvidenceAnchors: []ReplayEvidenceAnchor{result.PrimaryEvidence},
				})
			}
		}
	}

	// ----- Z-score population anomaly dimension -----
	if opt.ZScore.Enabled {
		zScore, hit := rse.computeZScoreDimension(context, opt.ZScore)
		if hit {
			weight := opt.ZScore.Weight
			if weight > 0 {
				if context.AccountAgeDays < 7 {
					weight *= 1.5
				}
				result.Dimensions["zscore_anomaly"] = zScore
				result.EffectiveWeights["zscore_anomaly"] = weight
				contribution := zScore * weight
				weightedTotal += contribution
				totalWeight += weight
				result.IndicatorDetails = append(result.IndicatorDetails, RiskIndicatorDetail{
					Name:            "zscore_anomaly",
					RawValue:        zScore,
					NormalizedScore: clampRiskScore(zScore),
					Weight:          weight,
					Contribution:    contribution,
					Explanation:     "population z-score anomaly",
					EvidenceAnchors: []ReplayEvidenceAnchor{result.PrimaryEvidence},
				})
			}
		}
	}

	return weightedTotal, totalWeight
}

// computeZScoreDimension scores how far the player's indicators sit from the global
// population in standard-deviation units. Returns (score, hit) where hit is false
// when no indicator exceeds the configured threshold (so it adds no contribution).
func (rse *RiskScoringEngine) computeZScoreDimension(context *DetectionContext, cfg ZScoreConfig) (float64, bool) {
	if len(context.GlobalBaselines) == 0 {
		return 0, false
	}
	maxAbsZ := 0.0
	checked := func(indicator string, observed float64) {
		if g, ok := context.GlobalBaselines[indicator]; ok && g.StdDev > 0 {
			z := math.Abs(observed-g.Mean) / g.StdDev
			if z > maxAbsZ {
				maxAbsZ = z
			}
		}
	}
	if len(context.ResponseTimes) > 0 {
		checked(baselineIndicatorResponseMean, meanInt64(context.ResponseTimes))
	}
	if context.HasWinRate {
		checked(baselineIndicatorWinRate, context.WinRate)
	}

	if maxAbsZ < cfg.Threshold {
		return 0, false
	}
	return zToScore(maxAbsZ), true
}

// applyRiskDecay folds the player's decayed historical risk into the current score.
// The decay is exponential in the number of consecutive normal games since the last
// violation, with a wall-clock time floor so rapid games cannot bypass it. Disabled
// by default. The caller must hold configLock.
func (rse *RiskScoringEngine) applyRiskDecay(context *DetectionContext, result *RiskScoringResult) {
	cfg := rse.config.Optimization.RiskDecay
	if !cfg.Enabled || context.HistoricalRisk <= 0 {
		return
	}

	now := context.Now
	if now.IsZero() {
		now = time.Now()
	}

	// Game-driven exponential decay.
	factor := math.Pow(cfg.DecayFactor, float64(context.NormalGamesSinceViolation))

	// Time floor: decay cannot complete faster than the configured minimum elapsed
	// time since the last violation. While inside the floor window, clamp the factor
	// so historical risk cannot be fully shed by rapid games.
	if cfg.MinFloorHours > 0 && context.LastViolationAt != nil {
		elapsed := now.Sub(*context.LastViolationAt)
		floor := time.Duration(cfg.MinFloorHours) * time.Hour
		if elapsed < floor {
			// Linear floor: at t=0 no decay (factor=1), approaching the floor lets the
			// game-driven factor take over. Use the larger (less-decayed) factor.
			ratio := elapsed.Seconds() / floor.Seconds()
			floorFactor := 1.0 - ratio // 1 -> 0 across the window
			if floorFactor > factor {
				factor = floorFactor
			}
		}
	}

	if factor < 0 {
		factor = 0
	}
	if factor > 1 {
		factor = 1
	}

	decayed := context.HistoricalRisk * factor
	f := factor
	result.DecayFactorApplied = &f

	// Blend the decayed historical contribution into the current score. Use a
	// max-style combine so history can raise but the current evidence still leads.
	if decayed > 0 {
		blended := result.RiskScore + decayed*0.5
		result.RiskScore = clampRiskScore(blended)
		result.IndicatorDetails = append(result.IndicatorDetails, RiskIndicatorDetail{
			Name:            "historical_risk_decayed",
			RawValue:        context.HistoricalRisk,
			NormalizedScore: clampRiskScore(decayed),
			Weight:          0.5,
			Contribution:    decayed * 0.5,
			Explanation:     "decayed historical risk folded into current score",
			EvidenceAnchors: []ReplayEvidenceAnchor{result.PrimaryEvidence},
		})
	}
}

// applyNewPlayerProtection relaxes the risk score for accounts in the observation
// period and marks the result so the sanction layer can suppress automatic bans.
//
// Multi-account / IP-clustering signals are NOT relaxed: their weighted
// contribution is preserved at full strength and only the remaining score is
// scaled by the relaxation factor. This prevents smurf accounts from being
// shielded by the observation period. The caller must hold configLock.
func (rse *RiskScoringEngine) applyNewPlayerProtection(context *DetectionContext, result *RiskScoringResult) {
	cfg := rse.config.Optimization.NewPlayer
	if !cfg.Enabled || !context.IsNewPlayer {
		return
	}
	result.NewPlayerObserve = true
	if cfg.RelaxationFactor < 0 || cfg.RelaxationFactor >= 1 {
		return
	}

	// Isolate the multi-account / IP contribution so it bypasses relaxation.
	protectedScore := multiAccountContribution(result)
	relaxable := result.RiskScore - protectedScore
	if relaxable < 0 {
		relaxable = 0
	}
	result.RiskScore = clampRiskScore(relaxable*cfg.RelaxationFactor + protectedScore)
}

// multiAccountSignals lists the dimension names treated as multi-account / IP
// clustering evidence that must not be relaxed by new-player protection.
var multiAccountSignals = map[string]bool{
	"multi_account": true,
	"ip_cluster":    true,
}

// multiAccountContribution returns the portion of the (normalized) risk score
// attributable to multi-account signals, so it can be excluded from relaxation.
func multiAccountContribution(result *RiskScoringResult) float64 {
	if len(result.IndicatorDetails) == 0 {
		return 0
	}
	totalWeight := 0.0
	for _, w := range result.EffectiveWeights {
		totalWeight += w
	}
	if totalWeight <= 0 {
		return 0
	}
	protected := 0.0
	for _, d := range result.IndicatorDetails {
		if multiAccountSignals[d.Name] {
			protected += d.Contribution / totalWeight
		}
	}
	if protected < 0 {
		return 0
	}
	return protected
}
