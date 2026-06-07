package yaps

import (
	"reflect"
	"testing"
)

func TestParseHashtags(t *testing.T) {
	got := parseHashtags("Люблю #Go и #go, а ещё #web3!")
	want := []string{"go", "web3"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("parseHashtags = %v, want %v", got, want)
	}
}

func TestParseMentions(t *testing.T) {
	got := parseMentions("привет @neo и @Neo и @trinity_1")
	want := []string{"neo", "trinity_1"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("parseMentions = %v, want %v", got, want)
	}
}
