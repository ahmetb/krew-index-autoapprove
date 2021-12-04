package bump

import "testing"

func TestIsNewPluginSubmission(t *testing.T) {
	cases := []struct {
		tc   testCase
		want bool
	}{
		{tc: NewPluginSubmission,
			want: true},
		{tc: MultipleNewPluginSubmissions,
			want: true},
		{tc: MultiFileChangeButNotYAML,
			want: false},
		{tc: SinglePluginBumpNonTrivial,
			want: false},
		{tc: SinglePluginBumpTrivial,
			want: false},
	}
	for _, c := range cases {
		t.Run(c.tc.name, func(tt *testing.T) {
			got, err := IsNewPluginSubmission(downloadPatch(tt, c.tc.patchURL))
			if err != nil {
				tt.Fatal(err)
			}
			if got != c.want {
				tt.Fatalf("got:%v want:%v", got, c.want)
			}
		})
	}
}
