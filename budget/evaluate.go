package budget

import "fmt"

func Evaluate(request Request) Decision {
	decision := Decision{
		Schema:         "gooo/meta-budget/decision/v1",
		Status:         StatusRefuted,
		NextOperation:  "STOP_FAIL_CLOSED",
		RequestedPlan:  request.Operation,
		BudgetExceeded: []string{},
		Pair:           PairSummary{Regressions: []MetricDelta{}, Deltas: []MetricDelta{}},
		BaselineAfter:  request.Baseline.After.Cost,
		CandidateAfter: request.Candidate.After.Cost,
	}

	if err := request.validateStatic(); err != nil {
		return refuted(decision, "VALIDATION", "VALIDATE_REQUEST", "MALFORMED_REQUEST", err.Error())
	}
	decision.MetaMetrics = append([]Metric{}, request.MetaMetrics...)

	if request.CacheHit {
		cachedDigest, err := request.Candidate.Before.Identity.Digest()
		if err != nil {
			return refuted(decision, "CACHE", "VERIFY_CACHE_IDENTITY", "CACHE_IDENTITY_UNAVAILABLE", err.Error())
		}
		if request.CachedIdentityDigest != cachedDigest {
			return refuted(decision, "CACHE", "VERIFY_CACHE_IDENTITY", "CACHE_IDENTITY_MISMATCH", "cached identity digest does not match recomputed exact identity")
		}
	}

	baselineExact, baselineErr := exactPairState(request.Baseline, "baseline")
	if baselineErr != nil {
		return refuted(decision, "PAIRING", "VERIFY_BASELINE_PAIR", "BASELINE_PAIR_INVALID", baselineErr.Error())
	}
	candidateExact, candidateErr := exactPairState(request.Candidate, "candidate")
	if candidateErr != nil {
		return refuted(decision, "PAIRING", "VERIFY_CANDIDATE_PAIR", "CANDIDATE_PAIR_INVALID", candidateErr.Error())
	}
	decision.Pair.ExactBaselinePair = baselineExact
	decision.Pair.ExactCandidatePair = candidateExact

	if violations := request.Budget.violations(request.Candidate.After.Cost); len(violations) > 0 {
		proposal, proposalErr := makeProposal(request.Operation, violations, request.Operation)
		if proposalErr != nil {
			return refuted(decision, "BUDGET", "GENERATE_ALTERNATIVE_PROPOSAL", "ALTERNATIVE_PLAN_INVALID", proposalErr.Error())
		}
		decision.Status = StatusUnknown
		decision.NextOperation = "REVIEW_ALLOWED_EXECUTION_PLAN_PROPOSAL"
		decision.BudgetExceeded = violations
		decision.Proposal = &proposal
		decision.Unknown = &UnknownRecord{
			Stage:         "BUDGET",
			Step:          "GENERATE_ALTERNATIVE_PROPOSAL",
			Reason:        fmt.Sprintf("candidate exceeded budget: %v", violations),
			UnknownClass:  "BUDGET_EXCEEDED",
			NextOperation: "REVIEW_ALLOWED_EXECUTION_PLAN_PROPOSAL",
			BlockedBy:     "budget_contract",
		}
		return decision
	}

	if !baselineExact || !candidateExact {
		return unknown(decision, "PAIRING", "VERIFY_EXACT_BEFORE_AFTER_PAIR", "EXACT_PAIR_MISSING", "COLLECT_EXACT_BEFORE_AFTER_PAIR", "input/tool/contract identity evidence")
	}
	if !samePairIdentity(request.Baseline, request.Candidate) {
		return unknown(decision, "PAIRING", "VERIFY_PAIRED_WORKLOAD", "PAIRED_WORKLOAD_MISSING", "COLLECT_SAME_WORKLOAD_BASELINE_CANDIDATE", "same input/tool/contract digest")
	}
	decision.Pair.SameWorkload = request.Baseline.Before.Identity.WorkloadID == request.Candidate.Before.Identity.WorkloadID
	decision.Pair.SameInputToolContract = sameInputToolContract(request.Baseline, request.Candidate)
	if !decision.Pair.SameWorkload || !decision.Pair.SameInputToolContract {
		return unknown(decision, "PAIRING", "VERIFY_PAIRED_WORKLOAD", "PAIRED_WORKLOAD_MISSING", "COLLECT_SAME_WORKLOAD_BASELINE_CANDIDATE", "same input/tool/contract digest")
	}

	decision.Pair.Deltas = metricDelta(request.Baseline.After.Cost, request.Candidate.After.Cost)
	for _, delta := range decision.Pair.Deltas {
		if delta.Delta > 0 {
			decision.Pair.Regressions = append(decision.Pair.Regressions, delta)
		}
	}
	if len(decision.Pair.Regressions) > 0 {
		return refuted(decision, "REGRESSION", "COMPARE_PAIRED_COST", "CANDIDATE_SLOWER_OR_LARGER", "paired candidate cost is worse and is reported, not hidden")
	}

	decision.Status = StatusClosed
	decision.Pair.ImprovementProven = hasStrictImprovement(decision.Pair.Deltas)
	decision.Pair.UtilityInferred = false
	if decision.Pair.ImprovementProven {
		decision.NextOperation = "SELECT_NEXT_META_OPERATION_FROM_PAIRED_EVIDENCE"
	} else {
		decision.NextOperation = "PRESERVE_PLAN_WITHOUT_UTILITY_INFERENCE"
	}
	return decision
}

func sameInputToolContract(a, b BeforeAfter) bool {
	return a.Before.Identity.InputDigest == b.Before.Identity.InputDigest &&
		a.After.Identity.InputDigest == b.After.Identity.InputDigest &&
		a.Before.Identity.ToolDigest == b.Before.Identity.ToolDigest &&
		a.After.Identity.ToolDigest == b.After.Identity.ToolDigest &&
		a.Before.Identity.ContractDigest == b.Before.Identity.ContractDigest &&
		a.After.Identity.ContractDigest == b.After.Identity.ContractDigest
}

func hasStrictImprovement(deltas []MetricDelta) bool {
	for _, delta := range deltas {
		if delta.Delta < 0 {
			return true
		}
	}
	return false
}

func makeProposal(requested ExecutionPlan, violations []string, source ExecutionPlan) (Proposal, error) {
	proposal := Proposal{
		Reason:                       fmt.Sprintf("budget exceeded for %v", violations),
		PreservedSemanticResolution: requested.SemanticResolution,
		PreservedProofDenominator:   requested.ProofDenominator,
		Alternatives:                []ExecutionPlan{},
	}
	for _, operation := range source.Operations {
		alternative := ExecutionPlan{
			ID:                 "proposal." + operation,
			Kind:               "STANDARD",
			SemanticResolution: requested.SemanticResolution,
			ProofDenominator:   requested.ProofDenominator,
			PrivilegeLevel:     "read-only",
			Operations:         []string{operation},
			Allowed:            true,
		}
		if err := alternative.Validate(&requested); err != nil {
			return Proposal{}, err
		}
		proposal.Alternatives = append(proposal.Alternatives, alternative)
	}
	if len(proposal.Alternatives) == 0 {
		return Proposal{}, fmt.Errorf("no allowed alternative operations")
	}
	return proposal, nil
}

func unknown(decision Decision, stage, step, class, next, blocked string) Decision {
	decision.Status = StatusUnknown
	decision.NextOperation = next
	decision.Unknown = &UnknownRecord{
		Stage:         stage,
		Step:          step,
		Reason:        class,
		UnknownClass:  class,
		NextOperation: next,
		BlockedBy:     blocked,
	}
	return decision
}

func refuted(decision Decision, stage, step, reason, detail string) Decision {
	decision.Status = StatusRefuted
	decision.NextOperation = "STOP_FAIL_CLOSED"
	decision.Refutation = &Refutation{
		Stage:     stage,
		Step:      step,
		Reason:    reason + ": " + detail,
		BlockedBy: step,
	}
	return decision
}
