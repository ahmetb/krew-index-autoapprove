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
