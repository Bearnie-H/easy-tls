package plugins

import (
	"fmt"
	"strconv"
	"strings"
)

// SemanticVersion represents a useable semantic version number.  This can be used to assert compatability between plugins, agents, and frameworks.
type SemanticVersion struct {
	MajorRelease int
	MinorRelease int
	Build        int
}

func (v *SemanticVersion) String() string {
	return fmt.Sprintf("%d.%d.%d", v.MajorRelease, v.MinorRelease, v.Build)
}

// ParseVersion allows for a SemanticVersion to be recovered from its string representation.
func ParseVersion(v string) (*SemanticVersion, error) {
	fields := strings.Split(v, ".")
	if len(fields) != 3 {
		return nil, fmt.Errorf("semantic version parse error - Got %d expected 3 (%s)", len(fields), v)
	}
	var (
		version = &SemanticVersion{}
		err     error
	)

	version.MajorRelease, err = strconv.Atoi(fields[0])
	if err != nil {
		return nil, fmt.Errorf("semantic version parse error - Expected integer, got %s - %s", fields[0], err)
	}

	version.MinorRelease, err = strconv.Atoi(fields[1])
	if err != nil {
		return nil, fmt.Errorf("semantic version parse error - Expected integer, got %s - %s", fields[1], err)
	}

	version.Build, err = strconv.Atoi(fields[2])
	if err != nil {
		return nil, fmt.Errorf("semantic version parse error - Expected integer, got %s - %s", fields[2], err)
	}

	return version, nil
}

// Accepts determines if a given version is accepted for a set of minimum and maximum acceptable versions
func Accepts(Is, AcceptsMin, AcceptsMax SemanticVersion) bool {

	if Is.MajorRelease < AcceptsMin.MajorRelease || Is.MajorRelease > AcceptsMax.MajorRelease {
		return false
	}

	// If bounds are the same, minor release is what matters.
	if AcceptsMin.MajorRelease == AcceptsMax.MajorRelease {
		return Is.MinorRelease >= AcceptsMin.MinorRelease && Is.MinorRelease <= AcceptsMax.MinorRelease
	}

	// If this is the same minor release as the minimum, make sure the minor release is higher
	if Is.MajorRelease == AcceptsMin.MajorRelease {
		return Is.MinorRelease >= AcceptsMin.MinorRelease
	}

	// If this is the same minor release as the maximum, make sure the minor release is lower
	if Is.MajorRelease == AcceptsMax.MajorRelease {
		return Is.MinorRelease <= AcceptsMax.MinorRelease
	}

	// Otherwise, the major release is between the two values, to any minor release is good.
	return true
}
