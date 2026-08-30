package budget

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

func InventoryRoot(root string) (Inventory, error) {
	root, err := filepath.Abs(root)
	if err != nil {
		return Inventory{}, err
	}
	info, err := os.Stat(root)
	if err != nil {
		return Inventory{}, err
	}
	if !info.IsDir() {
		return Inventory{}, fmt.Errorf("inventory root is not a directory: %s", root)
	}

	result := Inventory{RootReadmeExcluded: true, Metrics: []Metric{}}
	err = filepath.WalkDir(root, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if path != root && entry.IsDir() && entry.Name() == ".git" {
			return filepath.SkipDir
		}
		if path == root || entry.IsDir() {
			if path != root {
				result.DescendantDirs++
			}
			return nil
		}
		if !entry.Type().IsRegular() {
			return nil
		}
		relative, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		if relative == "README.md" {
			return nil
		}
		result.RegularFiles++
		extension := strings.ToLower(filepath.Ext(path))
		if extension != ".go" && extension != ".gooo" {
			return nil
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		lines := physicalLines(data)
		if extension == ".go" {
			result.GoFiles++
			result.GoLines += lines
		} else {
			result.GoooFiles++
			result.GoooLines += lines
		}
		return nil
	})
	if err != nil {
		return Inventory{}, err
	}
	result.Metrics = []Metric{
		inventoryMetric("repository.descendant_dirs", result.DescendantDirs, "FOUNDATION"),
		inventoryMetric("repository.regular_files", result.RegularFiles, "FOUNDATION"),
		inventoryMetric("source.go.files", result.GoFiles, "COHERENCE"),
		inventoryMetric("source.go.lines", result.GoLines, "COHERENCE"),
		inventoryMetric("source.gooo.files", result.GoooFiles, "DRIVER"),
		inventoryMetric("source.gooo.lines", result.GoooLines, "DRIVER"),
	}
	return result, nil
}

func physicalLines(data []byte) int64 {
	if len(data) == 0 {
		return 0
	}
	lines := int64(1)
	for _, byteValue := range data {
		if byteValue == '\n' {
			lines++
		}
	}
	if data[len(data)-1] == '\n' {
		lines--
	}
	return lines
}

func inventoryMetric(id string, value int64, class string) Metric {
	return Metric{
		ID:             id,
		Classification: ProofClass(class),
		Numerator:      value,
		Denominator:    1,
		Binding: MetricBinding{
			MetricID:          id,
			MetaActivityID:    "gooo.meta_budget.inventory." + id,
			SourceID:          "source.meta_budget.inventory",
			IRID:              "ir.meta_budget.inventory.v1",
			GeneratedArtifact: "artifact.meta_budget.inventory.v1",
			EvaluatorID:       "evaluator.meta_budget.inventory.v1",
		},
	}
}
