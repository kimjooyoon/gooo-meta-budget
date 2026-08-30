package budget

import "fmt"

func BuildReport(decision Decision, inventory Inventory, build, test, conformance PhaseMeasurement, conformanceRun ConformanceResult, artifactFiles, artifactBytes, executedTests, reusedTests, skippedTests, repositoryWrites int64) (ArtifactReport, error) {
	if build.Phase != "build" || test.Phase != "test" || conformance.Phase != "conformance" {
		return ArtifactReport{}, fmt.Errorf("phase measurements must be build, test, and conformance")
	}
	if artifactFiles < 0 || artifactBytes < 0 || executedTests < 0 || reusedTests < 0 || skippedTests < 0 || repositoryWrites < 0 {
		return ArtifactReport{}, fmt.Errorf("artifact, test, and write counts must be non-negative")
	}
	report := ArtifactReport{
		Schema:           "gooo/meta-budget/artifact/v1",
		Decision:         decision,
		Inventory:        inventory,
		Build:            build,
		Test:             test,
		Conformance:      conformance,
		ExecutedTests:    executedTests,
		ReusedTests:      reusedTests,
		SkippedTests:     skippedTests,
		ArtifactFiles:    artifactFiles,
		ArtifactBytes:    artifactBytes,
		RepositoryWrites: repositoryWrites,
		ConformanceRun:   conformanceRun,
		StatusCounts:     StatusCounts{},
		MetricDenominators: map[string]int64{
			"proof":    int64(decision.RequestedPlan.ProofDenominator),
			"physical": 1,
		},
		Metrics: append(append([]Metric{}, decision.MetaMetrics...), inventory.Metrics...),
	}
	report.StatusCounts = report.StatusCounts.Add(decision.Status)
	report.StatusCounts = report.StatusCounts.Add(conformanceRunStatus(conformanceRun))
	report.Metrics = append(report.Metrics,
		runtimeMetric("runtime.build.wall_ms", build.WallMS, Driver),
		runtimeMetric("runtime.test.wall_ms", test.WallMS, Driver),
		runtimeMetric("runtime.conformance.wall_ms", conformance.WallMS, Driver),
		runtimeMetric("runtime.peak_rss_kib", maxPhaseRSS(build, test, conformance), Guardrail),
		runtimeMetric("evidence.executed_tests", executedTests, Outcome),
		runtimeMetric("evidence.reused_tests", reusedTests, Driver),
		runtimeMetric("evidence.skipped_tests", skippedTests, Guardrail),
		runtimeMetric("artifact.files", artifactFiles, Outcome),
		runtimeMetric("artifact.bytes", artifactBytes, Outcome),
		runtimeMetric("repository.writes", repositoryWrites, Guardrail),
	)
	for _, metric := range report.Metrics {
		if err := metric.Validate(0); err != nil {
			return ArtifactReport{}, err
		}
	}
	return report, nil
}

func conformanceRunStatus(result ConformanceResult) Status {
	if result.Refuted > 0 {
		return StatusRefuted
	}
	if result.Unknown > 0 {
		return StatusUnknown
	}
	return StatusClosed
}

func maxPhaseRSS(measurements ...PhaseMeasurement) int64 {
	var peak int64
	for _, measurement := range measurements {
		if measurement.PeakRSSKiB > peak {
			peak = measurement.PeakRSSKiB
		}
	}
	return peak
}

func runtimeMetric(id string, value int64, class ProofClass) Metric {
	return Metric{
		ID:             id,
		Classification: class,
		Numerator:      value,
		Denominator:    1,
		Binding: MetricBinding{
			MetricID:          id,
			MetaActivityID:    "gooo.meta_budget.runtime." + id,
			SourceID:          "source.meta_budget.runtime.v1",
			IRID:              "ir.meta_budget.runtime.v1",
			GeneratedArtifact: "artifact.meta_budget.runtime.v1",
			EvaluatorID:       "evaluator.meta_budget.runtime.v1",
		},
	}
}
