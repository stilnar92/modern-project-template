package config

import "testing"

func TestString(t *testing.T) {
	t.Setenv("CONFIG_TEST", "")
	if got := String("CONFIG_TEST", "x"); got != "x" {
		t.Fatalf("got %q, want the fallback", got)
	}
	t.Setenv("CONFIG_TEST", "y")
	if got := String("CONFIG_TEST", "x"); got != "y" {
		t.Fatalf("got %q, want y", got)
	}
}

func TestRequired(t *testing.T) {
	t.Setenv("CONFIG_TEST", "")
	if _, err := Required("CONFIG_TEST"); err == nil {
		t.Fatal("want an error for an empty variable")
	}
	t.Setenv("CONFIG_TEST", "y")
	if v, err := Required("CONFIG_TEST"); err != nil || v != "y" {
		t.Fatalf("got %q, %v", v, err)
	}
}
