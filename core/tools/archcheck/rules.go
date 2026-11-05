package main

import (
	"fmt"
	"math"
	"sort"
	"strings"
)

// internalPrefix is the import path prefix of all internal packages.
const internalPrefix = "github.com/stilnar92/modern-project-template/core/internal/"

// Package is one Go package and the packages that it imports.
type Package struct {
	Path    string
	Imports []string
	// TestImports are the imports of the _test.go files. The import rules
	// apply to them too. The coupling report does not count them.
	TestImports []string
}

// Violation is one broken architecture rule.
type Violation struct {
	From string
	To   string
	Rule string
}

func (v Violation) String() string {
	return fmt.Sprintf("%s imports %s: %s", v.From, v.To, v.Rule)
}

// location tells which module owns a package, and whether the package is
// the public root of the module or a private subpackage.
type location struct {
	module  string // for example "catalog", "product/reports", "platform/health"
	private bool   // true for a subpackage below the module root
	product bool   // true for a product module
}

// locate returns the location of an internal package. ok is false for a
// package outside core/internal (for example, core/cmd/api).
func locate(path string) (loc location, ok bool) {
	rel, found := strings.CutPrefix(path, internalPrefix)
	if !found {
		return location{}, false
	}
	parts := strings.Split(rel, "/")
	switch {
	case (parts[0] == "product" || parts[0] == "platform") && len(parts) >= 2:
		loc.module = parts[0] + "/" + parts[1]
		loc.private = len(parts) > 2
		loc.product = parts[0] == "product"
	default:
		loc.module = parts[0]
		loc.private = len(parts) > 1
	}
	return loc, true
}

// CheckImports applies the import rules of backend standard 01.
func CheckImports(pkgs []Package) []Violation {
	var out []Violation
	for _, p := range pkgs {
		from, fromInternal := locate(p.Path)
		for _, imp := range append(p.Imports[:len(p.Imports):len(p.Imports)], p.TestImports...) {
			to, ok := locate(imp)
			if !ok {
				continue
			}
			if fromInternal && !from.product && to.product {
				out = append(out, Violation{p.Path, imp, "the core must not import a product module"})
				continue
			}
			sameModule := fromInternal && from.module == to.module
			if to.private && !sameModule {
				out = append(out, Violation{p.Path, imp, "use the root package of module " + to.module + ", not its internal subpackage"})
			}
		}
	}
	return out
}

// Coupling is the module coupling report.
type Coupling struct {
	Modules  int
	Links    int
	KMean    float64 // K̄: the mean number of incoming links per module
	KStdDev  float64
	InDegree map[string]int
	Hotspots []string // modules with an in-degree above K̄ + σ
}

// CouplingReport counts the links between modules from the import graph.
func CouplingReport(pkgs []Package) Coupling {
	modules := map[string]bool{}
	edges := map[[2]string]bool{}
	for _, p := range pkgs {
		from, ok := locate(p.Path)
		if !ok {
			continue
		}
		modules[from.module] = true
		for _, imp := range p.Imports {
			to, ok := locate(imp)
			if !ok || to.module == from.module {
				continue
			}
			modules[to.module] = true
			edges[[2]string{from.module, to.module}] = true
		}
	}
	in := map[string]int{}
	for m := range modules {
		in[m] = 0
	}
	for e := range edges {
		in[e[1]]++
	}
	c := Coupling{Modules: len(modules), Links: len(edges), InDegree: in}
	if c.Modules == 0 {
		return c
	}
	sum := 0
	for _, k := range in {
		sum += k
	}
	c.KMean = float64(sum) / float64(c.Modules)
	variance := 0.0
	for _, k := range in {
		d := float64(k) - c.KMean
		variance += d * d
	}
	c.KStdDev = math.Sqrt(variance / float64(c.Modules))
	for m, k := range in {
		if float64(k) > c.KMean+c.KStdDev {
			c.Hotspots = append(c.Hotspots, m)
		}
	}
	sort.Strings(c.Hotspots)
	return c
}
