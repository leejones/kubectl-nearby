package output_test

import (
	"strings"
	"testing"
	"time"

	"github.com/leejones/kubectl-nearby/pkg/output"
)

func TestAge(t *testing.T) {
	var testCases = []struct {
		input  string
		output string
	}{
		{"5s", "5s"},
		{"119s", "119s"},
		// >= 120s then use minutes and seconds
		{"121s", "2m1s"},
		{"9m59s", "9m59s"},
		// >= 10m, use minutes
		{"10m", "10m"},
		{"119m59s", "119m"},
		// >= 120m, use hours
		{"120m", "2h"},
		{"23h59m59s", "23h"},
		// >= 1d, use days
		{"24h0m1s", "1d"},
	}
	for _, testCase := range testCases {
		input, err := time.ParseDuration(testCase.input)
		if err != nil {
			t.Errorf("Unexpected error parsing testCase intput: %v, error: %v", testCase.input, err)
		}
		want := testCase.output
		got := output.Age(input)
		if want != got {
			t.Errorf("Expected ageOutput(%v) to return: %v, but got: %v", input, want, got)
		}
	}
}

func TestColumns(t *testing.T) {
	want := strings.Trim(`
NAMESPACE   NAME               READY
default     foo-bar-abc123     1/3  
production  baz-bat-db-def456  2/2  
`, "\n")
	input := [][]string{
		{"NAMESPACE", "NAME", "READY"},
		{"default", "foo-bar-abc123", "1/3"},
		{"production", "baz-bat-db-def456", "2/2"},
	}
	got, err := output.Columns(input)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if want != got {
		t.Errorf("Expected ColumnOutput to return:\n%v\n--- but got: ---\n%v", want, got)
	}
}
