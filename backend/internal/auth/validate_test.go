package auth

import "testing"

func TestValidateRegister(t *testing.T) {
	cases := []struct {
		name    string
		in      RegisterInput
		wantErr bool
		field   string
	}{
		{"ok", RegisterInput{Username: "neo_99", Email: "neo@matrix.io", DisplayName: "Neo", Password: "trinity8"}, false, ""},
		{"bad username", RegisterInput{Username: "no", Email: "a@b.co", Password: "password1"}, true, "username"},
		{"bad email", RegisterInput{Username: "valid_one", Email: "nope", Password: "password1"}, true, "email"},
		{"short password", RegisterInput{Username: "valid_one", Email: "a@b.co", Password: "short"}, true, "password"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			err := ValidateRegister(c.in)
			if c.wantErr != (err != nil) {
				t.Fatalf("wantErr=%v got %v", c.wantErr, err)
			}
			if c.wantErr {
				verr, ok := err.(*ValidationError)
				if !ok {
					t.Fatalf("expected *ValidationError, got %T", err)
				}
				if verr.Field != c.field {
					t.Fatalf("field: want %q got %q", c.field, verr.Field)
				}
			}
		})
	}
}

func TestNormalizeLocale(t *testing.T) {
	for in, want := range map[string]string{"ru": "ru", "RU-ru": "ru", "en": "en", "fr": "en", "": "en"} {
		if got := NormalizeLocale(in); got != want {
			t.Errorf("NormalizeLocale(%q)=%q want %q", in, got, want)
		}
	}
}
