package config

import "testing"

func TestAnyProfileStoresToken(t *testing.T) {
	t.Parallel()
	if anyProfileStoresToken(map[string]CoreConfig{"a": {}}) {
		t.Fatal("expected false for empty token map")
	}
	if anyProfileStoresToken(map[string]CoreConfig{"a": {Token: "  "}}) {
		t.Fatal("expected false for whitespace-only token")
	}
	if !anyProfileStoresToken(map[string]CoreConfig{"a": {Token: "x"}}) {
		t.Fatal("expected true when a profile has a token")
	}
	if !anyProfileStoresToken(map[string]CoreConfig{"a": {}, "b": {Token: "y"}}) {
		t.Fatal("expected true when any profile has a token")
	}
}
