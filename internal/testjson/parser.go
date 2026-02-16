package testjson

import (
	"bufio"
	"encoding/json"
	"io"
	"strings"
)

// packageBuilder accumulates events for a single package during parsing.
type packageBuilder struct {
	pkg   string
	tests map[string]*TestResult
	order []string // insertion order of test names
	output []string
}

func newPackageBuilder(pkg string) *packageBuilder {
	return &packageBuilder{
		pkg:   pkg,
		tests: make(map[string]*TestResult),
	}
}

func (b *packageBuilder) addTest(name string) {
	if _, exists := b.tests[name]; !exists {
		b.tests[name] = &TestResult{Name: name}
		b.order = append(b.order, name)
	}
}

func (b *packageBuilder) finalize(elapsed float64, action Action) *PackageResult {
	result := &PackageResult{
		Package:       b.pkg,
		Elapsed:       elapsed,
		Output:        b.output,
		PackageAction: action,
	}

	for _, name := range b.order {
		result.Tests = append(result.Tests, b.tests[name])
	}

	// Detect special package states from output.
	for _, line := range b.output {
		if strings.Contains(line, "[build failed]") {
			result.BuildFailed = true
		}
		if strings.Contains(line, "[no test files]") {
			result.NoTestFiles = true
		}
	}

	return result
}

// Parse reads a go test -json stream and returns the aggregated results per package.
func Parse(r io.Reader) ([]*PackageResult, error) {
	scanner := bufio.NewScanner(r)

	builders := make(map[string]*packageBuilder)
	var order []string // package insertion order
	var results []*PackageResult

	for scanner.Scan() {
		line := scanner.Text()

		var event TestEvent
		if err := json.Unmarshal([]byte(line), &event); err != nil {
			// Non-JSON line: attach to most recent package output.
			if len(order) > 0 {
				last := order[len(order)-1]
				builders[last].output = append(builders[last].output, line)
			}
			continue
		}

		pkg := event.Package
		if pkg == "" && event.ImportPath != "" {
			// Go 1.24+ emits build-output/build-fail events with ImportPath
			// instead of Package. Derive the package name from the import path.
			pkg = strings.SplitN(event.ImportPath, " ", 2)[0]
		}
		if pkg == "" {
			continue
		}

		b, exists := builders[pkg]
		if !exists {
			b = newPackageBuilder(pkg)
			builders[pkg] = b
			order = append(order, pkg)
		}

		switch event.Action {
		case ActionRun:
			if event.Test != "" {
				b.addTest(event.Test)
			}

		case ActionOutput:
			if event.Test != "" {
				b.addTest(event.Test)
				b.tests[event.Test].Output = append(b.tests[event.Test].Output, event.Output)
			} else {
				b.output = append(b.output, event.Output)
			}

		case ActionPass:
			if event.Test != "" {
				b.addTest(event.Test)
				b.tests[event.Test].Status = StatusPass
			} else {
				// Package-level pass: finalize.
				results = append(results, b.finalize(event.Elapsed, ActionPass))
				delete(builders, pkg)
			}

		case ActionFail:
			if event.Test != "" {
				b.addTest(event.Test)
				b.tests[event.Test].Status = StatusFail
			} else {
				// Package-level fail: finalize.
				results = append(results, b.finalize(event.Elapsed, ActionFail))
				delete(builders, pkg)
			}

		case ActionSkip:
			if event.Test != "" {
				b.addTest(event.Test)
				b.tests[event.Test].Status = StatusSkip
			} else {
				// Package-level skip: finalize.
				results = append(results, b.finalize(event.Elapsed, ActionSkip))
				delete(builders, pkg)
			}

		case ActionBuildOutput:
			b.output = append(b.output, event.Output)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return results, nil
}
