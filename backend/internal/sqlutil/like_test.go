package sqlutil

import "testing"

func TestContainsPattern_escapesWildcards(t *testing.T) {
	got := ContainsPattern(`100%_off`)
	want := `%100\%\_off%`
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}
