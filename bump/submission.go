package bump

import (
	"fmt"
	"strings"

	"github.com/sourcegraph/go-diff/diff"
)

func IsNewPluginSubmission(patch []byte) (bool, error) {
	fileDiffs, err := diff.ParseMultiFileDiff(patch)
	if err != nil {
		return false, err
	}
	if len(fileDiffs) == 0 {
		return false, err
	}
	for _, file := range fileDiffs {
		if !(strings.HasSuffix(file.NewName, ".yaml") &&
			file.OrigName == "/dev/null" &&
			file.Stat().Deleted == 0) {
			return false, nil
		}
	}
	return true, nil
}

func IsReviewablePluginSubmission(patch []byte) error {
	fileDiffs, err := diff.ParseMultiFileDiff(patch)
	if err != nil {
		return fmt.Errorf("internal parse error: %w", err)
	}
	for len(fileDiffs) > 1 {
		return fmt.Errorf("please submit only one new plugin in a pull request as each plugin is evaluated and accepted independently")
	}
	return nil
}
