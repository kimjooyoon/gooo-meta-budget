package budget

import (
	"path/filepath"
	"testing"
)

func loadCase(t *testing.T, name string) ConformanceCase {
	t.Helper()
	var testCase ConformanceCase
	path := filepath.Join("..", "fixtures", "cases", name+".json")
	if err := ReadJSON(path, &testCase); err != nil {
		t.Fatal(err)
	}
	return testCase
}

func TestNormalEvidenceCanProveCostImprovement(t *testing.T) {
	decision := Evaluate(loadCase(t, "normal").Request)
	if decision.Status != StatusClosed {
		t.Fatalf("status = %s, want %s", decision.Status, StatusClosed)
	}
	if !decision.Pair.ExactBaselinePair || !decision.Pair.ExactCandidatePair || !decision.Pair.ImprovementProven {
		t.Fatalf("exact paired improvement was not proven: %+v", decision.Pair)
	}
	if decision.Pair.UtilityInferred {
		t.Fatal("utility must never be inferred from cost evidence")
	}
}

func TestBudgetProposalPreservesResolutionAndDenominator(t *testing.T) {
	decision := Evaluate(loadCase(t, "budget-exceeded").Request)
	if decision.Status != StatusUnknown || decision.Proposal == nil {
		t.Fatalf("budget decision = %+v, want UNKNOWN proposal", decision)
	}
	if decision.Proposal.PreservedSemanticResolution != "full" || decision.Proposal.PreservedProofDenominator != 6 {
		t.Fatalf("proposal lowered proof contract: %+v", decision.Proposal)
	}
	for _, alternative := range decision.Proposal.Alternatives {
		if alternative.SemanticResolution != "full" || alternative.ProofDenominator != 6 || alternative.PrivilegeLevel != "read-only" || !alternative.Allowed {
			t.Fatalf("unsafe alternative: %+v", alternative)
		}
	}
}

func TestCacheHitDoesNotCloseMissingPair(t *testing.T) {
	decision := Evaluate(loadCase(t, "cache-hit").Request)
	if decision.Status != StatusUnknown {
		t.Fatalf("status = %s, want UNKNOWN", decision.Status)
	}
	if decision.Unknown == nil || decision.Unknown.UnknownClass != "EXACT_PAIR_MISSING" {
		t.Fatalf("unknown record = %+v", decision.Unknown)
	}
}

func TestInvalidInputsFailClosed(t *testing.T) {
	for _, name := range []string{"malformed", "fixed-point", "privilege-escalation", "refuted", "regression"} {
		decision := Evaluate(loadCase(t, name).Request)
		if decision.Status != StatusRefuted {
			t.Errorf("%s status = %s, want REFUTED", name, decision.Status)
		}
	}
}

func TestStatusPrecedence(t *testing.T) {
	if got := AggregateStatus([]Status{StatusClosed, StatusUnknown}); got != StatusUnknown {
		t.Fatalf("closed + unknown = %s", got)
	}
	if got := AggregateStatus([]Status{StatusClosed, StatusUnknown, StatusRefuted}); got != StatusRefuted {
		t.Fatalf("closed + unknown + refuted = %s", got)
	}
}
