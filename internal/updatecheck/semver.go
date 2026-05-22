package updatecheck

import (
	"fmt"
	"strconv"
	"strings"
)

type Semver struct {
	Major int
	Minor int
	Patch int
}

func ParseSemver(s string) (Semver, error) {
	s = strings.TrimSpace(s)
	s = strings.TrimPrefix(s, "v")

	if i := strings.IndexAny(s, "+-"); i >= 0 {
		s = s[:i]
	}

	parts := strings.Split(s, ".")
	if len(parts) < 3 {
		return Semver{}, fmt.Errorf("invalid semver %q", s)
	}

	major, err := strconv.Atoi(parts[0])
	if err != nil {
		return Semver{}, fmt.Errorf("invalid semver %q", s)
	}
	minor, err := strconv.Atoi(parts[1])
	if err != nil {
		return Semver{}, fmt.Errorf("invalid semver %q", s)
	}
	patch, err := strconv.Atoi(parts[2])
	if err != nil {
		return Semver{}, fmt.Errorf("invalid semver %q", s)
	}

	return Semver{Major: major, Minor: minor, Patch: patch}, nil
}

func (v Semver) LessThan(other Semver) bool {
	if v.Major != other.Major {
		return v.Major < other.Major
	}
	if v.Minor != other.Minor {
		return v.Minor < other.Minor
	}
	return v.Patch < other.Patch
}
