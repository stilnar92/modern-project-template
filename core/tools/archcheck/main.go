// Command archcheck checks the architecture rules of backend standard 01:
// the import rules and table ownership. It also prints the coupling report.
//
// Run it from the core folder: go run ./tools/archcheck
package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"sort"
	"strings"
)

func main() {
	pkgs, err := listPackages(context.Background())
	if err != nil {
		fmt.Fprintln(os.Stderr, "archcheck:", err)
		os.Exit(2)
	}
	failed := false

	for _, v := range CheckImports(pkgs) {
		fmt.Println("import rule:", v)
		failed = true
	}

	tables, err := CheckTables("internal")
	if err != nil {
		fmt.Fprintln(os.Stderr, "archcheck:", err)
		os.Exit(2)
	}
	for _, v := range tables {
		fmt.Printf("table ownership: %s %s: %s\n", v.File, v.Table, v.Note)
		failed = true
	}

	printCoupling(CouplingReport(pkgs))

	if failed {
		os.Exit(1)
	}
	fmt.Println("archcheck: all rules pass")
}

// listPackages reads the import graph with "go list".
func listPackages(ctx context.Context) ([]Package, error) {
	cmd := exec.CommandContext(ctx, "go", "list", "-json", "./...")
	var out, stderr bytes.Buffer
	cmd.Stdout, cmd.Stderr = &out, &stderr
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("go list: %w: %s", err, stderr.String())
	}
	var pkgs []Package
	dec := json.NewDecoder(&out)
	for {
		var p struct {
			ImportPath   string
			Imports      []string
			TestImports  []string
			XTestImports []string
		}
		err := dec.Decode(&p)
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return nil, err
		}
		pkgs = append(pkgs, Package{
			Path:        p.ImportPath,
			Imports:     p.Imports,
			TestImports: append(p.TestImports, p.XTestImports...),
		})
	}
	return pkgs, nil
}

func printCoupling(c Coupling) {
	fmt.Printf("coupling report: %d modules, %d links, K̄ = %.2f (target 1–3), σ = %.2f\n",
		c.Modules, c.Links, c.KMean, c.KStdDev)
	names := make([]string, 0, len(c.InDegree))
	for m := range c.InDegree {
		names = append(names, m)
	}
	sort.Strings(names)
	for _, m := range names {
		fmt.Printf("  %-28s incoming links: %d\n", m, c.InDegree[m])
	}
	if len(c.Hotspots) > 0 {
		fmt.Println("  hotspots:", strings.Join(c.Hotspots, ", "))
	}
}
