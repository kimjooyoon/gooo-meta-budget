package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/kimjooyoon/gooo-meta-budget/budget"
)

func main() {
	log.SetFlags(0)
	if len(os.Args) < 2 {
		log.Fatal("command is required: evaluate, inventory, measure, conformance, or report")
	}
	var err error
	switch os.Args[1] {
	case "evaluate":
		err = evaluateCommand(os.Args[2:])
	case "inventory":
		err = inventoryCommand(os.Args[2:])
	case "measure":
		err = measureCommand(os.Args[2:])
	case "conformance":
		err = conformanceCommand(os.Args[2:])
	case "report":
		err = reportCommand(os.Args[2:])
	default:
		err = fmt.Errorf("unknown command: %s", os.Args[1])
	}
	if err != nil {
		log.Fatal(err)
	}
}

func evaluateCommand(args []string) error {
	flags := flag.NewFlagSet("evaluate", flag.ContinueOnError)
	input := flags.String("input", "", "request JSON path")
	output := flags.String("output", "", "decision JSON path")
	if err := flags.Parse(args); err != nil {
		return err
	}
	if *input == "" || *output == "" {
		return errors.New("evaluate requires --input and --output")
	}
	var request budget.Request
	if err := budget.ReadJSON(*input, &request); err != nil {
		return err
	}
	return budget.WriteJSON(*output, budget.Evaluate(request))
}

func inventoryCommand(args []string) error {
	flags := flag.NewFlagSet("inventory", flag.ContinueOnError)
	root := flags.String("root", ".", "repository root")
	output := flags.String("output", "", "inventory JSON path")
	if err := flags.Parse(args); err != nil {
		return err
	}
	if *output == "" {
		return errors.New("inventory requires --output")
	}
	inventory, err := budget.InventoryRoot(*root)
	if err != nil {
		return err
	}
	return budget.WriteJSON(*output, inventory)
}

func measureCommand(args []string) error {
	flags := flag.NewFlagSet("measure", flag.ContinueOnError)
	phase := flags.String("phase", "", "build, test, or conformance")
	output := flags.String("output", "", "measurement JSON path")
	if err := flags.Parse(args); err != nil {
		return err
	}
	command := flags.Args()
	if len(command) > 0 && command[0] == "--" {
		command = command[1:]
	}
	_, err := budget.MeasureCommand(*phase, *output, command)
	return err
}

func conformanceCommand(args []string) error {
	flags := flag.NewFlagSet("conformance", flag.ContinueOnError)
	cases := flags.String("cases", "fixtures/cases", "conformance case directory")
	output := flags.String("output", "", "conformance JSON path")
	if err := flags.Parse(args); err != nil {
		return err
	}
	if *output == "" {
		return errors.New("conformance requires --output")
	}
	result, runErr := budget.RunConformance(*cases)
	if err := budget.WriteJSON(*output, result); err != nil {
		return err
	}
	return runErr
}

func reportCommand(args []string) error {
	flags := flag.NewFlagSet("report", flag.ContinueOnError)
	decisionPath := flags.String("decision", "", "decision JSON path")
	inventoryPath := flags.String("inventory", "", "inventory JSON path")
	buildPath := flags.String("build", "", "build measurement JSON path")
	testPath := flags.String("test", "", "test measurement JSON path")
	conformancePath := flags.String("conformance", "", "conformance measurement JSON path")
	runPath := flags.String("conformance-run", "", "conformance result JSON path")
	artifactRoot := flags.String("artifact-root", "", "generated artifact directory")
	output := flags.String("output", "", "artifact report JSON path")
	executed := flags.Int64("executed-tests", -1, "executed test count")
	reused := flags.Int64("reused-tests", -1, "reused test count")
	skipped := flags.Int64("skipped-tests", -1, "skipped test count")
	writes := flags.Int64("repository-writes", -1, "repository write count")
	if err := flags.Parse(args); err != nil {
		return err
	}
	paths := []*string{decisionPath, inventoryPath, buildPath, testPath, conformancePath, runPath, artifactRoot, output}
	for _, path := range paths {
		if *path == "" {
			return errors.New("report requires decision, inventory, build, test, conformance, conformance-run, artifact-root, and output paths")
		}
	}
	var decision budget.Decision
	var inventory budget.Inventory
	var build, test, conformance budget.PhaseMeasurement
	var conformanceRun budget.ConformanceResult
	for path, value := range map[string]any{
		*decisionPath:    &decision,
		*inventoryPath:   &inventory,
		*buildPath:       &build,
		*testPath:        &test,
		*conformancePath: &conformance,
		*runPath:         &conformanceRun,
	} {
		if err := budget.ReadJSON(path, value); err != nil {
			return err
		}
	}
	artifactFiles, artifactBytes, err := budget.ArtifactSize(*artifactRoot)
	if err != nil {
		return err
	}
	executedCount := chooseCount(*executed, decision.CandidateAfter.ExecutedTests)
	reusedCount := chooseCount(*reused, decision.CandidateAfter.ReusedTests)
	skippedCount := chooseCount(*skipped, decision.CandidateAfter.SkippedTests)
	writesCount := chooseCount(*writes, decision.CandidateAfter.RepositoryWrites)
	report, err := budget.BuildReport(decision, inventory, build, test, conformance, conformanceRun, artifactFiles, artifactBytes, executedCount, reusedCount, skippedCount, writesCount)
	if err != nil {
		return err
	}
	return budget.WriteJSON(*output, report)
}

func chooseCount(value, fallback int64) int64 {
	if value >= 0 {
		return value
	}
	return fallback
}
