package budget

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

func RunConformance(directory string) (ConformanceResult, error) {
	entries, err := os.ReadDir(directory)
	if err != nil {
		return ConformanceResult{}, err
	}
	paths := make([]string, 0, len(entries))
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".json") {
			paths = append(paths, filepath.Join(directory, entry.Name()))
		}
	}
	sort.Strings(paths)
	result := ConformanceResult{Schema: "gooo/meta-budget/conformance/v1", Cases: []CaseResult{}}
	seen := make(map[string]bool, len(paths))
	for _, path := range paths {
		var testCase ConformanceCase
		if err := ReadJSON(path, &testCase); err != nil {
			return result, fmt.Errorf("read conformance case %s: %w", path, err)
		}
		if testCase.Name == "" || seen[testCase.Name] {
			return result, fmt.Errorf("conformance case name is missing or duplicated: %s", path)
		}
		seen[testCase.Name] = true
		decision := Evaluate(testCase.Request)
		passed := decision.Status == testCase.ExpectedStatus
		if decision.Status == StatusUnknown && !hasCompleteUnknown(decision.Unknown) {
			passed = false
		}
		caseResult := CaseResult{Name: testCase.Name, ExpectedStatus: testCase.ExpectedStatus, Status: decision.Status, Passed: passed}
		result.Total++
		if passed {
			result.Passed++
		}
		switch decision.Status {
		case StatusClosed:
			result.Closed++
		case StatusUnknown:
			result.Unknown++
		case StatusRefuted:
			result.Refuted++
		}
		result.Cases = append(result.Cases, caseResult)
	}
	required := []string{"normal", "unknown", "refuted", "malformed", "fixed-point", "privilege-escalation", "cache-hit", "cache-mismatch", "budget-exceeded", "regression"}
	for _, name := range required {
		if !seen[name] {
			return result, fmt.Errorf("required conformance case is missing: %s", name)
		}
	}
	if result.Passed != result.Total {
		return result, fmt.Errorf("conformance failed: %d/%d cases passed", result.Passed, result.Total)
	}
	return result, nil
}

func hasCompleteUnknown(record *UnknownRecord) bool {
	return record != nil && record.Stage != "" && record.Step != "" && record.Reason != "" && record.UnknownClass != "" && record.NextOperation != "" && record.BlockedBy != ""
}
