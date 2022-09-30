package bump

import (
	"testing"
)

func TestIsBumpPatch(t *testing.T) {
	cases := []struct {
		tc   testCase
		want bool
	}{
		{tc: NewPluginSubmission,
			want: false},
		{tc: MultipleNewPluginSubmissions,
			want: false},
		{tc: MultiFileChangeButNotYAML,
			want: false},
		{tc: SinglePluginBumpNonTrivial,
			want: true},
		{tc: SinglePluginBumpTrivial,
			want: true},
		{tc: SinglePluginSameArtifactMultipleOS,
			want: true},
	}
	for _, c := range cases {
		t.Run(c.tc.name, func(tt *testing.T) {
			got, err := IsBumpPatch(downloadPatch(tt, c.tc.patchURL))
			if err != nil {
				tt.Fatal(err)
			}
			if got != c.want {
				tt.Fatalf("got:%v want:%v", got, c.want)
			}
		})
	}
}

func TestIsValidBump(t *testing.T) {
	cases := []struct {
		tc          testCase
		expectError bool
	}{
		{tc: SinglePluginSameArtifactMultipleOS},
		{tc: SinglePluginBumpTrivial},
		{tc: SinglePluginBumpNonTrivial, expectError: true},
	}
	for _, c := range cases {
		t.Run(c.tc.name, func(tt *testing.T) {
			err := IsValidBump(downloadPatch(tt, c.tc.patchURL))
			if !c.expectError && err != nil {
				tt.Fatalf("got unexpected error %v", err)
			}

			if c.expectError && err == nil {
				tt.Fatalf("expected error, got none")
			}
		})
	}
}
