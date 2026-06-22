package ids

import "testing"

func TestParseAndString(t *testing.T) {
	t.Parallel()

	raw := "task_01JY4F4EG7Y0M9M8Y1SSQJ5VTR"
	id, err := Parse(raw)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if id.String() != raw {
		t.Fatalf("got %q want %q", id.String(), raw)
	}
	if err := Validate(raw, PrefixTask); err != nil {
		t.Fatalf("validate: %v", err)
	}
}

func TestValidateWrongPrefix(t *testing.T) {
	t.Parallel()

	err := Validate("evt_01JY4F4EG7Y0M9M8Y1SSQJ5VTR", PrefixTask)
	if err == nil {
		t.Fatal("expected error for wrong prefix")
	}
}

func TestParseInvalid(t *testing.T) {
	t.Parallel()

	cases := []string{"", "nounderscore", "task_", "_value"}
	for _, c := range cases {
		if _, err := Parse(c); err == nil {
			t.Fatalf("expected error for %q", c)
		}
	}
}
