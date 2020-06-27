package plugins

import (
	"fmt"
)

// SemanticVersion represents a useable semantic version number.  This can be used to assert compatability between plugins, agents, and frameworks.
type SemanticVersion struct {
	MajorRelease int
	MinorRelease int
	Build        int
}

const (
	versionFmtString string = "%d.%d.%d"
)

func (v *SemanticVersion) String() string {
	return fmt.Sprintf(versionFmtString, v.MajorRelease, v.MinorRelease, v.Build)
}

// ParseVersion allows for a SemanticVersion to be recovered from its string representation.
func ParseVersion(v string) (*SemanticVersion, error) {

	Version := &SemanticVersion{}

	n, err := fmt.Sscanf(v, versionFmtString, &Version.MajorRelease, &Version.MinorRelease, &Version.Build)
	if err != nil {
		return nil, err
	} else if n != 3 {
		return nil, fmt.Errorf("version error: Failed to parse string [ %s ] as version specifier", v)
	}

	return Version, nil
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

	// If this is the same major release as the minimum, make sure the minor release is higher
	if Is.MajorRelease == AcceptsMin.MajorRelease {
		return Is.MinorRelease >= AcceptsMin.MinorRelease
	}

	// If this is the same major release as the maximum, make sure the minor release is lower
	if Is.MajorRelease == AcceptsMax.MajorRelease {
		return Is.MinorRelease <= AcceptsMax.MinorRelease
	}

	// Otherwise, the major release is between the two values, so any minor release is good.
	return true
}
