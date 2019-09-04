package bump

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/sourcegraph/go-diff/diff"
	"k8s.io/apimachinery/pkg/util/version"
	"log"
	"regexp"
	"strings"
)

func IsBumpPatch(patch []byte) (bool, error) {
	fileDiffs, err := diff.ParseMultiFileDiff(patch)
	if err != nil {
		return false, err
	}
	var validFiles int
	for _, v := range fileDiffs {
		oldName := strings.TrimPrefix(v.OrigName, "a/")
		newName := strings.TrimPrefix(v.NewName, "b/")

		if oldName != newName {
			return false, nil
		}
		if !strings.HasSuffix(newName, ".yaml") {
			return false, nil
		}
		validFiles++
	}
	return validFiles > 0, nil
}

func IsValidBump(patch []byte) error {
	diffs, err := diff.ParseMultiFileDiff(patch)
	if err != nil {
		return fmt.Errorf("failed to parse diff: %v", err)
	}
	// TODO do nothing for newly added plugins

	// ensure file names
	for _, v := range diffs {
		oldName := strings.TrimPrefix(v.OrigName, "a/")
		newName := strings.TrimPrefix(v.NewName, "b/")

		if oldName != newName {
			return fmt.Errorf("file name changed (%q --> %q)", oldName, newName)
		}
		if !strings.HasSuffix(newName, ".yaml") {
			return fmt.Errorf("a file doesn't have .yaml suffix: %s", newName)
		}

		if err := isBumpPlugin(v); err != nil {
			return fmt.Errorf("file %s is not a straightforward version bump: %v", newName, err)
		}
	}

	return nil
}

func isBumpPlugin(d *diff.FileDiff) error {
	vA, vB, ok := findVersionSpecs(d)
	if !ok {
		return errors.New("could not find the old/new version spec in the diff")
	}

	svA, err := version.ParseSemantic(vA)
	if err != nil {
		return fmt.Errorf("could not parse version string %q", vA)
	}
	svB, err := version.ParseSemantic(vB)
	if err != nil {
		return fmt.Errorf("could not parse version string %q", vB)
	}

	if !svA.LessThan(svB) {
		return fmt.Errorf("version should move forward (%q vs %q)", vA, vB)
	}

	log.Printf("oldVersion: %s, newVersion: %s", svA, svB)

	var urlChanges bool
	for _, hunk := range d.Hunks {
		ok, err := isBumpHunk(hunk.Body, vA, vB)
		if err != nil {
			return err
		}
		urlChanges = urlChanges || ok
	}

	if !urlChanges {
		return errors.New("no 'uri:' field changes done in the patch")
	}

	return nil
}

var (
	diffLine        = regexp.MustCompile(`^[\+\-]`)
	versionDiffLine = regexp.MustCompile(`^[\+\-]\s+version:\s`)
	oldURLDiffLine  = regexp.MustCompile(`(?m)^\-\s+(\- )?uri:\s(.*)`)
	newURLDiffLine  = regexp.MustCompile(`(?m)^\+\s+(\- )?uri:\s(.*)`)
	sumDiffLine     = regexp.MustCompile(`^[\+\-]\s+(\- )?sha256:\s(.*)`)
)

func isBumpHunk(hunk []byte, vA, vB string) (bool, error) {
	lines := bytes.Split(hunk, []byte{'\n'})

	var hasURL bool

	for _, line := range lines {
		if !diffLine.Match(line) {
			continue
		}

		if versionDiffLine.Match(line) || sumDiffLine.Match(line) {
			continue
		}
		if oldURLDiffLine.Match(line) || newURLDiffLine.Match(line) {
			hasURL = true
			continue
		}
		return false, fmt.Errorf("diff line unrecognized for version bumps: [%s]", string(line))
	}

	if hasURL {
		ua, ub, ok := findURLSpecs(hunk)
		if !ok {
			return false, errors.New("found changes to 'uri:' field(s) but can't find old/new url in the patch")
		}

		// sometimes people don't include v* prefix in file names
		vA = strings.TrimPrefix(vA, "v")
		vB = strings.TrimPrefix(vB, "v")

		uab := strings.ReplaceAll(ua, vA, vB)
		if uab != ub {
			return false, fmt.Errorf("changing old version (%q) with new version (%q) in the url (%s) did not result in the new url (%s), expected: %s", vA, vB, ua, ub, uab)
		}
	}
	return hasURL, nil
}

var (
	oldVersionSpec = regexp.MustCompile(`-\s+version:\s?(.+)`)
	newVersionSpec = regexp.MustCompile(`\+\s+version:\s?(.+)`)
)

func findVersionSpecs(d *diff.FileDiff) (string, string, bool) {
	for _, hunk := range d.Hunks {
		vOld := oldVersionSpec.FindSubmatch(hunk.Body)
		vNew := newVersionSpec.FindSubmatch(hunk.Body)
		if len(vOld) >= 2 || len(vNew) >= 2 {
			return strings.Trim(string(vOld[len(vOld)-1]), `"`),
				strings.Trim(string(vNew[len(vNew)-1]), `"`), true
		}
	}
	return "", "", false
}

func findURLSpecs(hunk []byte) (string, string, bool) {
	ua := oldURLDiffLine.FindSubmatch(hunk)
	ub := newURLDiffLine.FindSubmatch(hunk)

	if len(ua) < 2 || len(ub) < 2 {
		return "", "", false
	}

	return string(bytes.Trim(ua[len(ua)-1], `"`)),
		string(bytes.Trim(ub[len(ub)-1], `"`)), true
}
