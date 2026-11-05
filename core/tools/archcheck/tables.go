package main

import (
	"bufio"
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	pg_query "github.com/pganalyze/pg_query_go/v6"
)

// tablesLine is the line in a module README that lists the tables of the
// module, for example: "Tables: services, works".
var tablesLine = regexp.MustCompile(`(?i)^\s*tables:\s*(.*)$`)

// OwnedTables reads the "Tables:" line of a module README.
func OwnedTables(readme string) (tables map[string]bool, found bool) {
	f, err := os.Open(readme)
	if err != nil {
		return nil, false
	}
	defer func() { _ = f.Close() }()
	s := bufio.NewScanner(f)
	for s.Scan() {
		m := tablesLine.FindStringSubmatch(s.Text())
		if m == nil {
			continue
		}
		tables = map[string]bool{}
		for _, t := range strings.Split(m[1], ",") {
			if t = strings.TrimSpace(strings.Trim(strings.TrimSpace(t), "`")); t != "" && t != "-" {
				tables[strings.ToLower(t)] = true
			}
		}
		return tables, true
	}
	return nil, false
}

// UsedTables returns the tables that a SQL text refers to. It uses the
// Postgres parser (the same one that sqlc uses), so upserts, CTEs and
// quoted names work. A reference to a CTE in its scope is not a table.
func UsedTables(sql string) ([]string, error) {
	tree, err := pg_query.ParseToJSON(sql)
	if err != nil {
		return nil, err
	}
	var root any
	if err := json.Unmarshal([]byte(tree), &root); err != nil {
		return nil, err
	}
	tables := map[string]bool{}
	walk(root, nil, tables)
	out := make([]string, 0, len(tables))
	for t := range tables {
		out = append(out, t)
	}
	sort.Strings(out)
	return out, nil
}

// walk collects the table references (RangeVar nodes). ctes holds the CTE
// names that are visible at this point of the tree.
func walk(node any, ctes map[string]bool, tables map[string]bool) {
	switch n := node.(type) {
	case []any:
		for _, v := range n {
			walk(v, ctes, tables)
		}
	case map[string]any:
		// A RangeVar is a table reference. Some fields (for example the
		// target of INSERT) hold it without the "RangeVar" key.
		if name, ok := n["relname"].(string); ok {
			if _, ok := n["relpersistence"]; ok {
				_, qualified := n["schemaname"]
				if qualified || !ctes[name] {
					tables[name] = true
				}
			}
		}
		// The target of INSERT, UPDATE, DELETE and MERGE is always a real
		// table, also when a CTE has the same name.
		if rel, ok := n["relation"].(map[string]any); ok {
			if name, ok := rel["relname"].(string); ok {
				tables[name] = true
			}
		}
		if with, ok := n["withClause"].(map[string]any); ok {
			ctes = walkWith(with, ctes, tables)
		}
		for k, v := range n {
			if k != "withClause" {
				walk(v, ctes, tables)
			}
		}
	}
}

// walkWith walks the CTE definitions of one WITH clause. A CTE body sees
// the CTEs before it. In WITH RECURSIVE, each body sees all CTEs. It returns the
// names that the rest of the statement sees.
func walkWith(with map[string]any, outer map[string]bool, tables map[string]bool) map[string]bool {
	recursive, _ := with["recursive"].(bool)
	visible := map[string]bool{}
	for k := range outer {
		visible[k] = true
	}
	list, _ := with["ctes"].([]any)
	var defs []map[string]any
	for _, item := range list {
		wrap, _ := item.(map[string]any)
		cte, _ := wrap["CommonTableExpr"].(map[string]any)
		defs = append(defs, cte)
		if recursive {
			visible[cte["ctename"].(string)] = true
		}
	}
	for _, cte := range defs {
		name, _ := cte["ctename"].(string)
		walk(cte["ctequery"], visible, tables)
		next := map[string]bool{name: true}
		for k := range visible {
			next[k] = true
		}
		visible = next
	}
	return visible
}

// TableViolation is a query that uses a table that its module does not own.
type TableViolation struct {
	File  string
	Table string
	Note  string
}

// CheckTables checks each module folder that has a queries/ folder. The
// sqlc queries of a module must use only the tables in its README.
func CheckTables(internalDir string) ([]TableViolation, error) {
	files, err := filepath.Glob(filepath.Join(internalDir, "*", "queries", "*.sql"))
	if err != nil {
		return nil, err
	}
	more, err := filepath.Glob(filepath.Join(internalDir, "*", "*", "queries", "*.sql"))
	if err != nil {
		return nil, err
	}
	files = append(files, more...)
	var out []TableViolation
	for _, file := range files {
		moduleDir := filepath.Dir(filepath.Dir(file))
		owned, found := OwnedTables(filepath.Join(moduleDir, "README.md"))
		if !found {
			out = append(out, TableViolation{File: file, Note: "the module README has no \"Tables:\" line"})
			continue
		}
		data, err := os.ReadFile(file)
		if err != nil {
			return nil, err
		}
		used, err := UsedTables(string(data))
		if err != nil {
			out = append(out, TableViolation{File: file, Note: "the SQL does not parse: " + err.Error()})
			continue
		}
		for _, t := range used {
			if !owned[t] {
				out = append(out, TableViolation{File: file, Table: t, Note: "the module does not own this table"})
			}
		}
	}
	return out, nil
}
