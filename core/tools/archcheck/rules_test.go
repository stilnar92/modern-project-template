package main

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

const p = internalPrefix

func TestCheckImports(t *testing.T) {
	tests := []struct {
		name string
		pkg  Package
		want int
	}{
		{"core imports a product module", Package{Path: p + "catalog", Imports: []string{p + "product/reports"}}, 1},
		{"product module imports the core", Package{Path: p + "product/reports", Imports: []string{p + "catalog"}}, 0},
		{"module imports another module's subpackage", Package{Path: p + "orders", Imports: []string{p + "catalog/store"}}, 1},
		{"module imports its own subpackage", Package{Path: p + "catalog", Imports: []string{p + "catalog/store"}}, 0},
		{"cmd imports a module root", Package{Path: "github.com/stilnar92/modern-project-template/core/cmd/api", Imports: []string{p + "catalog"}}, 0},
		{"cmd imports a private subpackage", Package{Path: "github.com/stilnar92/modern-project-template/core/cmd/api", Imports: []string{p + "catalog/store"}}, 1},
		{"module imports a platform root", Package{Path: p + "catalog", Imports: []string{p + "platform/health"}}, 0},
		{"module imports a platform subpackage", Package{Path: p + "catalog", Imports: []string{p + "platform/db/internal"}}, 1},
		{"test imports another module's subpackage", Package{Path: p + "orders", TestImports: []string{p + "catalog/store"}}, 1},
		{"standard library import", Package{Path: p + "catalog", Imports: []string{"net/http"}}, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := CheckImports([]Package{tt.pkg}); len(got) != tt.want {
				t.Fatalf("got %d violations %v, want %d", len(got), got, tt.want)
			}
		})
	}
}

func TestCouplingReport(t *testing.T) {
	pkgs := []Package{
		{Path: p + "catalog", Imports: []string{p + "accounts"}},
		{Path: p + "orders", Imports: []string{p + "accounts", p + "catalog"}},
		{Path: p + "sites", Imports: []string{p + "accounts"}},
		{Path: p + "accounts", Imports: nil},
	}
	c := CouplingReport(pkgs)
	if c.Modules != 4 || c.Links != 4 {
		t.Fatalf("got %d modules and %d links, want 4 and 4", c.Modules, c.Links)
	}
	if c.KMean != 1 {
		t.Fatalf("K̄ = %v, want 1", c.KMean)
	}
	if !reflect.DeepEqual(c.Hotspots, []string{"accounts"}) {
		t.Fatalf("hotspots = %v, want [accounts]", c.Hotspots)
	}
}

func TestUsedTables(t *testing.T) {
	tests := []struct {
		name string
		sql  string
		want []string
	}{
		{"select with join and schema", "SELECT s.id FROM public.products s JOIN variants w ON w.product_id = s.id", []string{"products", "variants"}},
		{"words in comments do not count", "-- from the comment table\nSELECT 1 FROM payments", []string{"payments"}},
		{"upsert", "INSERT INTO orders (id, name) VALUES ($1, $2) ON CONFLICT (id) DO UPDATE SET name = EXCLUDED.name", []string{"orders"}},
		{"cte name is not a table", "WITH recent AS (SELECT * FROM payments) SELECT * FROM recent", []string{"payments"}},
		{"cte scope ends with the statement", "WITH orders AS (SELECT 1) SELECT * FROM orders; SELECT * FROM orders", []string{"orders"}},
		{"schema-qualified name is a table", "WITH orders AS (SELECT 1) SELECT * FROM public.orders", []string{"orders"}},
		{"cte body sees the real table", "WITH orders AS (SELECT * FROM orders) SELECT * FROM orders", []string{"orders"}},
		{"recursive cte sees itself", "WITH RECURSIVE t AS (SELECT 1 AS n UNION ALL SELECT n + 1 FROM t) SELECT * FROM t", []string{}},
		{"write target is a table", "WITH orders AS (SELECT 1) DELETE FROM orders", []string{"orders"}},
		{"recursive forward reference", "WITH RECURSIVE a AS (SELECT * FROM b), b AS (SELECT 1) SELECT * FROM a", []string{}},
		{"quoted name", `SELECT * FROM "products"`, []string{"products"}},
		{"comma list", "SELECT * FROM products, orders", []string{"orders", "products"}},
		{"subquery and update", "UPDATE orders SET n = 1 WHERE id IN (SELECT order_id FROM payments)", []string{"orders", "payments"}},
		{"sqlc arguments", "-- name: Get :one\nSELECT * FROM orders WHERE id = sqlc.arg(id)", []string{"orders"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := UsedTables(tt.sql)
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("got %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCheckTables(t *testing.T) {
	dir := t.TempDir()
	write := func(rel, text string) {
		t.Helper()
		path := filepath.Join(dir, rel)
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(text), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	write("catalog/README.md", "# catalog\n\nTables: products, variants\n")
	write("catalog/queries/products.sql", "SELECT * FROM products JOIN variants ON true;")
	write("orders/README.md", "# orders\n\nTables: orders\n")
	write("orders/queries/orders.sql", `SELECT * FROM orders JOIN "products" ON true;`)
	write("product/reports/queries/c.sql", "SELECT * FROM reports;")

	got, err := CheckTables(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 {
		t.Fatalf("got %d violations %v, want 2 (orders uses products; reports has no README)", len(got), got)
	}
}
