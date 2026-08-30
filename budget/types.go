package budget

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
)

type Status string

const (
	StatusClosed  Status = "CLOSED"
	StatusUnknown Status = "UNKNOWN"
	StatusRefuted Status = "REFUTED"
)

type ProofClass string

const (
	Foundation ProofClass = "FOUNDATION"
	Coherence  ProofClass = "COHERENCE"
	Regression ProofClass = "REGRESSION"
	Driver     ProofClass = "DRIVER"
	Outcome    ProofClass = "OUTCOME"
	Guardrail  ProofClass = "GUARDRAIL"
)

var proofClasses = []ProofClass{
	Foundation,
	Coherence,
	Regression,
	Driver,
	Outcome,
	Guardrail,
}

type MetricBinding struct {
	MetricID          string `json:"metric_id"`
	MetaActivityID    string `json:"meta_activity_id"`
	SourceID          string `json:"source_id"`
	IRID              string `json:"ir_id"`
	GeneratedArtifact string `json:"generated_artifact_id"`
	EvaluatorID       string `json:"evaluator_id"`
}

type Metric struct {
	ID             string        `json:"id"`
	Classification ProofClass    `json:"classification"`
	Numerator      int64         `json:"numerator"`
	Denominator    int64         `json:"denominator"`
	Binding        MetricBinding `json:"binding"`
}

type Identity struct {
	WorkloadID    string `json:"workload_id"`
	InputDigest   string `json:"input_digest"`
	ToolDigest    string `json:"tool_digest"`
	ContractDigest string `json:"contract_digest"`
	IdentityDigest string `json:"identity_digest"`
}

type PhaseCost struct {
	WallMS     int64 `json:"wall_ms"`
	PeakRSSKiB int64 `json:"peak_rss_kib"`
}

type Cost struct {
	Build            PhaseCost `json:"build"`
	Test             PhaseCost `json:"test"`
	Conformance      PhaseCost `json:"conformance"`
	ExecutedTests    int64     `json:"executed_tests"`
	ReusedTests      int64     `json:"reused_tests"`
	SkippedTests     int64     `json:"skipped_tests"`
	ArtifactFiles    int64     `json:"artifact_files"`
	ArtifactBytes    int64     `json:"artifact_bytes"`
	RepositoryWrites int64     `json:"repository_writes"`
}

type Observation struct {
	Variant  string   `json:"variant"`
	Identity Identity `json:"identity"`
	Cost     Cost     `json:"cost"`
}

type BeforeAfter struct {
	Before Observation `json:"before"`
	After  Observation `json:"after"`
}

type ExecutionPlan struct {
	ID                 string   `json:"id"`
	Kind               string   `json:"kind"`
	SemanticResolution string   `json:"semantic_resolution"`
	ProofDenominator   int      `json:"proof_denominator"`
	PrivilegeLevel     string   `json:"privilege_level"`
	Operations         []string `json:"operations"`
	Allowed            bool     `json:"allowed"`
}

type Budget struct {
	BuildWallMS       int64 `json:"build_wall_ms"`
	TestWallMS        int64 `json:"test_wall_ms"`
	ConformanceWallMS int64 `json:"conformance_wall_ms"`
	PeakRSSKiB        int64 `json:"peak_rss_kib"`
	ArtifactBytes     int64 `json:"artifact_bytes"`
}

type Request struct {
	Operation             ExecutionPlan `json:"operation"`
	Budget                Budget        `json:"budget"`
	Baseline              BeforeAfter   `json:"baseline"`
	Candidate             BeforeAfter   `json:"candidate"`
	CacheHit              bool          `json:"cache_hit"`
	CachedIdentityDigest  string        `json:"cached_identity_digest"`
	MetaMetrics           []Metric      `json:"meta_metrics"`
}

type UnknownRecord struct {
	Stage         string `json:"stage"`
	Step          string `json:"step"`
	Reason        string `json:"reason"`
	UnknownClass  string `json:"unknown_class"`
	NextOperation string `json:"next_operation"`
	BlockedBy     string `json:"blocked_by"`
}

type Refutation struct {
	Stage     string `json:"stage"`
	Step      string `json:"step"`
	Reason    string `json:"reason"`
	BlockedBy string `json:"blocked_by"`
}

type MetricDelta struct {
	ID              string `json:"id"`
	Baseline        int64  `json:"baseline"`
	Candidate       int64  `json:"candidate"`
	Delta           int64  `json:"delta"`
	CandidateBetter bool   `json:"candidate_better"`
}

type PairSummary struct {
	ExactBaselinePair       bool          `json:"exact_baseline_pair"`
	ExactCandidatePair      bool          `json:"exact_candidate_pair"`
	SameWorkload            bool          `json:"same_workload"`
	SameInputToolContract   bool          `json:"same_input_tool_contract"`
	ImprovementProven       bool          `json:"improvement_proven"`
	UtilityInferred         bool          `json:"utility_inferred"`
	Regressions             []MetricDelta `json:"regressions"`
	Deltas                  []MetricDelta `json:"deltas"`
}

type Proposal struct {
	Reason                      string          `json:"reason"`
	PreservedSemanticResolution string          `json:"preserved_semantic_resolution"`
	PreservedProofDenominator   int             `json:"preserved_proof_denominator"`
	Alternatives                []ExecutionPlan `json:"alternatives"`
}

type Decision struct {
	Schema          string        `json:"schema"`
	Status          Status        `json:"status"`
	NextOperation   string        `json:"next_operation"`
	RequestedPlan   ExecutionPlan `json:"requested_plan"`
	BudgetExceeded  []string      `json:"budget_exceeded"`
	Proposal        *Proposal     `json:"proposal,omitempty"`
	Unknown         *UnknownRecord `json:"unknown,omitempty"`
	Refutation      *Refutation   `json:"refutation,omitempty"`
	Pair            PairSummary   `json:"pair"`
	BaselineAfter   Cost          `json:"baseline_after"`
	CandidateAfter  Cost          `json:"candidate_after"`
}

type StatusCounts struct {
	Closed  int `json:"closed"`
	Unknown int `json:"unknown"`
	Refuted int `json:"refuted"`
}

type CaseResult struct {
	Name           string `json:"name"`
	ExpectedStatus Status `json:"expected_status"`
	Status         Status `json:"status"`
	Passed         bool   `json:"passed"`
}

type ConformanceResult struct {
	Schema  string       `json:"schema"`
	Total   int          `json:"total"`
	Passed  int          `json:"passed"`
	Closed  int          `json:"closed"`
	Unknown int          `json:"unknown"`
	Refuted int          `json:"refuted"`
	Cases   []CaseResult `json:"cases"`
}

type ConformanceCase struct {
	Name           string  `json:"name"`
	ExpectedStatus Status  `json:"expected_status"`
	Request        Request `json:"request"`
}

type PhaseMeasurement struct {
	Phase      string `json:"phase"`
	ExitCode   int    `json:"exit_code"`
	WallMS     int64  `json:"wall_ms"`
	PeakRSSKiB int64  `json:"peak_rss_kib"`
}

type Inventory struct {
	RootReadmeExcluded bool     `json:"root_readme_excluded"`
	DescendantDirs     int64    `json:"descendant_dirs"`
	RegularFiles       int64    `json:"regular_files"`
	GoFiles            int64    `json:"go_files"`
	GoLines            int64    `json:"go_lines"`
	GoooFiles          int64    `json:"gooo_files"`
	GoooLines          int64    `json:"gooo_lines"`
	Metrics            []Metric `json:"metrics"`
}

type ArtifactReport struct {
	Schema           string              `json:"schema"`
	Decision         Decision             `json:"decision"`
	Inventory        Inventory            `json:"inventory"`
	Build            PhaseMeasurement    `json:"build"`
	Test             PhaseMeasurement    `json:"test"`
	Conformance      PhaseMeasurement    `json:"conformance"`
	ExecutedTests    int64               `json:"executed_tests"`
	ReusedTests      int64               `json:"reused_tests"`
	SkippedTests     int64               `json:"skipped_tests"`
	ArtifactFiles    int64               `json:"artifact_files"`
	ArtifactBytes    int64               `json:"artifact_bytes"`
	RepositoryWrites int64               `json:"repository_writes"`
	ConformanceRun   ConformanceResult   `json:"conformance_run"`
	StatusCounts     StatusCounts        `json:"status_counts"`
	MetricDenominators map[string]int64  `json:"metric_denominators"`
	Metrics          []Metric            `json:"metrics"`
}

func (i Identity) canonical() map[string]string {
	return map[string]string{
		"contract_digest": i.ContractDigest,
		"input_digest":    i.InputDigest,
		"tool_digest":     i.ToolDigest,
		"workload_id":     i.WorkloadID,
	}
}

func (i Identity) Digest() (string, error) {
	for key, value := range i.canonical() {
		if value == "" {
			return "", fmt.Errorf("identity %s is empty", key)
		}
	}
	payload, err := json.Marshal(i.canonical())
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(payload)
	return "sha256:" + hex.EncodeToString(sum[:]), nil
}

func (i Identity) Verify() error {
	expected, err := i.Digest()
	if err != nil {
		return err
	}
	if i.IdentityDigest != expected {
		return fmt.Errorf("identity digest mismatch: got %q want %q", i.IdentityDigest, expected)
	}
	return nil
}

func (c Cost) PeakRSS() int64 {
	peak := c.Build.PeakRSSKiB
	if c.Test.PeakRSSKiB > peak {
		peak = c.Test.PeakRSSKiB
	}
	if c.Conformance.PeakRSSKiB > peak {
		peak = c.Conformance.PeakRSSKiB
	}
	return peak
}

func (c Cost) values() map[string]int64 {
	return map[string]int64{
		"build.wall_ms":       c.Build.WallMS,
		"test.wall_ms":        c.Test.WallMS,
		"conformance.wall_ms": c.Conformance.WallMS,
		"peak_rss_kib":        c.PeakRSS(),
		"artifact.files":      c.ArtifactFiles,
		"artifact.bytes":      c.ArtifactBytes,
	}
}

func (b Budget) violations(c Cost) []string {
	violations := make([]string, 0, 5)
	if c.Build.WallMS > b.BuildWallMS {
		violations = append(violations, "build_wall_ms")
	}
	if c.Test.WallMS > b.TestWallMS {
		violations = append(violations, "test_wall_ms")
	}
	if c.Conformance.WallMS > b.ConformanceWallMS {
		violations = append(violations, "conformance_wall_ms")
	}
	if c.PeakRSS() > b.PeakRSSKiB {
		violations = append(violations, "peak_rss_kib")
	}
	if c.ArtifactBytes > b.ArtifactBytes {
		violations = append(violations, "artifact_bytes")
	}
	return violations
}

func validateNonNegative(name string, value int64) error {
	if value < 0 {
		return fmt.Errorf("%s must be non-negative", name)
	}
	return nil
}

func (p ExecutionPlan) Validate(requested *ExecutionPlan) error {
	if p.ID == "" || p.SemanticResolution == "" || p.PrivilegeLevel == "" {
		return errors.New("plan id, semantic resolution, and privilege level are required")
	}
	if p.Kind == "FIXED_POINT" {
		return errors.New("FIXED_POINT plans are not executable")
	}
	if p.PrivilegeLevel != "read-only" {
		return errors.New("privilege escalation is not allowed")
	}
	if p.ProofDenominator <= 0 {
		return errors.New("proof denominator must be positive")
	}
	if len(p.Operations) == 0 {
		return errors.New("plan must name at least one operation")
	}
	if requested != nil {
		if !p.Allowed {
			return errors.New("alternative plan is not allowed")
		}
		if p.SemanticResolution != requested.SemanticResolution {
			return errors.New("alternative lowers semantic resolution")
		}
		if p.ProofDenominator != requested.ProofDenominator {
			return errors.New("alternative lowers proof denominator")
		}
	}
	return nil
}

func (b Budget) Validate() error {
	if b.BuildWallMS <= 0 || b.TestWallMS <= 0 || b.ConformanceWallMS <= 0 || b.PeakRSSKiB <= 0 || b.ArtifactBytes <= 0 {
		return errors.New("all budget denominators must be positive")
	}
	return nil
}

func (m Metric) Validate(expectedDenominator int64) error {
	if m.ID == "" || m.Denominator <= 0 || m.Numerator < 0 {
		return errors.New("metric id, positive denominator, and non-negative numerator are required")
	}
	if expectedDenominator > 0 && m.Denominator != expectedDenominator {
		return fmt.Errorf("metric %s has denominator %d, want %d", m.ID, m.Denominator, expectedDenominator)
	}
	if m.Classification != Foundation && m.Classification != Coherence && m.Classification != Regression && m.Classification != Driver && m.Classification != Outcome && m.Classification != Guardrail {
		return fmt.Errorf("metric %s has unknown classification %q", m.ID, m.Classification)
	}
	if m.Binding.MetricID != m.ID || m.Binding.MetaActivityID == "" || m.Binding.SourceID == "" || m.Binding.IRID == "" || m.Binding.GeneratedArtifact == "" || m.Binding.EvaluatorID == "" {
		return fmt.Errorf("metric %s is not bound one-to-one to activity/source/IR/artifact/evaluator", m.ID)
	}
	return nil
}

func validateCost(c Cost) error {
	values := c.values()
	for name, value := range values {
		if err := validateNonNegative(name, value); err != nil {
			return err
		}
	}
	for name, value := range map[string]int64{
		"executed_tests": c.ExecutedTests,
		"reused_tests": c.ReusedTests,
		"skipped_tests": c.SkippedTests,
		"repository_writes": c.RepositoryWrites,
	} {
		if err := validateNonNegative(name, value); err != nil {
			return err
		}
	}
	if c.RepositoryWrites != 0 {
		return errors.New("repository writes must be zero")
	}
	return nil
}

func validateObservation(o Observation, expectedVariant string) error {
	if o.Variant != expectedVariant {
		return fmt.Errorf("observation variant %q, want %q", o.Variant, expectedVariant)
	}
	if err := o.Identity.Verify(); err != nil {
		return err
	}
	return validateCost(o.Cost)
}

func (r Request) validateStatic() error {
	if err := r.Operation.Validate(nil); err != nil {
		return err
	}
	if err := r.Budget.Validate(); err != nil {
		return err
	}
	seen := make(map[ProofClass]bool, len(proofClasses))
	for _, metric := range r.MetaMetrics {
		if err := metric.Validate(int64(r.Operation.ProofDenominator)); err != nil {
			return err
		}
		if seen[metric.Classification] {
			return fmt.Errorf("duplicate proof class %s", metric.Classification)
		}
		seen[metric.Classification] = true
	}
	if len(r.MetaMetrics) != len(proofClasses) {
		return fmt.Errorf("proof metrics must have denominator classes %d", len(proofClasses))
	}
	for _, class := range proofClasses {
		if !seen[class] {
			return fmt.Errorf("missing proof class %s", class)
		}
	}
	if r.CacheHit && r.CachedIdentityDigest == "" {
		return errors.New("cache hit requires cached identity digest")
	}
	return nil
}

func exactPairState(pair BeforeAfter, variant string) (bool, error) {
	if pair.Before.Identity.IdentityDigest == "" || pair.After.Identity.IdentityDigest == "" {
		return false, nil
	}
	if err := validateObservation(pair.Before, variant); err != nil {
		return false, err
	}
	if err := validateObservation(pair.After, variant); err != nil {
		return false, err
	}
	if pair.Before.Identity.IdentityDigest != pair.After.Identity.IdentityDigest {
		return false, nil
	}
	return true, nil
}

func samePairIdentity(a, b BeforeAfter) bool {
	return a.Before.Identity.IdentityDigest == b.Before.Identity.IdentityDigest &&
		a.After.Identity.IdentityDigest == b.After.Identity.IdentityDigest &&
		a.Before.Identity.WorkloadID == b.Before.Identity.WorkloadID &&
		a.After.Identity.WorkloadID == b.After.Identity.WorkloadID &&
		a.Before.Identity.InputDigest == b.Before.Identity.InputDigest &&
		a.After.Identity.InputDigest == b.After.Identity.InputDigest &&
		a.Before.Identity.ToolDigest == b.Before.Identity.ToolDigest &&
		a.After.Identity.ToolDigest == b.After.Identity.ToolDigest &&
		a.Before.Identity.ContractDigest == b.Before.Identity.ContractDigest &&
		a.After.Identity.ContractDigest == b.After.Identity.ContractDigest
}

func metricDelta(baseline, candidate Cost) []MetricDelta {
	baseValues := baseline.values()
	candidateValues := candidate.values()
	ids := []string{"build.wall_ms", "test.wall_ms", "conformance.wall_ms", "peak_rss_kib", "artifact.files", "artifact.bytes"}
	deltas := make([]MetricDelta, 0, len(ids))
	for _, id := range ids {
		base := baseValues[id]
		candidateValue := candidateValues[id]
		deltas = append(deltas, MetricDelta{ID: id, Baseline: base, Candidate: candidateValue, Delta: candidateValue - base, CandidateBetter: candidateValue < base})
	}
	return deltas
}

func statusRank(s Status) int {
	switch s {
	case StatusRefuted:
		return 2
	case StatusUnknown:
		return 1
	default:
		return 0
	}
}

func AggregateStatus(statuses []Status) Status {
	strongest := StatusClosed
	for _, status := range statuses {
		if statusRank(status) > statusRank(strongest) {
			strongest = status
		}
	}
	return strongest
}

func (s StatusCounts) Status() Status {
	if s.Refuted > 0 {
		return StatusRefuted
	}
	if s.Unknown > 0 {
		return StatusUnknown
	}
	return StatusClosed
}

func (s StatusCounts) Add(status Status) StatusCounts {
	switch status {
	case StatusRefuted:
		s.Refuted++
	case StatusUnknown:
		s.Unknown++
	default:
		s.Closed++
	}
	return s
}

func intString(value int64) string {
	return strconv.FormatInt(value, 10)
}
