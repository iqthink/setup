package steps

import (
	"reflect"
	"testing"
)

func TestTapsIn(t *testing.T) {
	tests := []struct {
		name string
		in   []string
		want []string
	}{
		{
			name: "no tap-qualified entries",
			in:   []string{"gum", "mise", "gh"},
			want: nil,
		},
		{
			name: "single tap, single package",
			in:   []string{"gum", "stripe/stripe-cli/stripe"},
			want: []string{"stripe/stripe-cli"},
		},
		{
			name: "multiple distinct taps preserve declaration order",
			in: []string{
				"gum",
				"stripe/stripe-cli/stripe",
				"basecamp/tap/basecamp-cli",
			},
			want: []string{"stripe/stripe-cli", "basecamp/tap"},
		},
		{
			name: "duplicate tap from two packages dedupes",
			in: []string{
				"some/tap/pkg-a",
				"some/tap/pkg-b",
			},
			want: []string{"some/tap"},
		},
		{
			name: "two-segment names are NOT treated as tap-qualified",
			in:   []string{"openssl@3", "node/18"},
			want: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tapsIn(tt.in)
			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("tapsIn(%v) = %v, want %v", tt.in, got, tt.want)
			}
		})
	}
}

func TestPipeline_RegistersTeamAppsSteps(t *testing.T) {
	// Guardrail: All() must include both TeamAppsMust and TeamAppsOptional
	// in that order (must before optional so a partial pipeline cancel
	// after must still gives the user a workable baseline).
	all := All()
	var seenMust, seenOptional int
	mustIdx, optIdx := -1, -1
	for i, s := range all {
		switch s.(type) {
		case TeamAppsMust:
			seenMust++
			mustIdx = i
		case TeamAppsOptional:
			seenOptional++
			optIdx = i
		}
	}
	if seenMust != 1 {
		t.Fatalf("TeamAppsMust appears %d times in All(), want 1", seenMust)
	}
	if seenOptional != 1 {
		t.Fatalf("TeamAppsOptional appears %d times in All(), want 1", seenOptional)
	}
	if mustIdx > optIdx {
		t.Fatalf("TeamAppsMust at index %d should precede TeamAppsOptional at %d", mustIdx, optIdx)
	}
}
