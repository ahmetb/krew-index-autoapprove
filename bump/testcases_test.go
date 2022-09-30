package bump

import (
	"io/ioutil"
	"net/http"
	"testing"
)

type testCase struct {
	name     string
	patchURL string
}

var (
	NewPluginSubmission = testCase{
		name:     "new plugin submission",
		patchURL: "https://github.com/kubernetes-sigs/krew-index/pull/1809.diff"}
	MultipleNewPluginSubmissions = testCase{
		name:     "multiple plugins added at once",
		patchURL: "https://github.com/kubernetes-sigs/krew-index/pull/631.diff"}
	MultiFileChangeButNotYAML = testCase{
		name:     "multi-file change but not plugin yaml",
		patchURL: "https://github.com/kubernetes-sigs/krew-index/pull/882.diff"}
	SinglePluginBumpNonTrivial = testCase{
		name:     "version bump but not straightforward",
		patchURL: "https://github.com/kubernetes-sigs/krew-index/pull/1791.diff"}
	SinglePluginBumpTrivial = testCase{
		name:     "version bump and straightforward",
		patchURL: "https://github.com/kubernetes-sigs/krew-index/pull/1808.diff"}
	SinglePluginSameArtifactMultipleOS = testCase{
		name:     "single plugin with same artifact for multiple os",
		patchURL: "https://github.com/kubernetes-sigs/krew-index/pull/2640.diff"}
)

func downloadPatch(t *testing.T, url string) []byte {
	t.Helper()
	r, err := http.Get(url)
	if err != nil {
		t.Fatal(err)
	}
	if r.StatusCode != http.StatusOK {
		t.Fatalf("status:%d", r.StatusCode)
	}
	defer r.Body.Close()
	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		t.Fatal(err)
	}
	return b
}
