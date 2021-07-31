package version

var (
	// Branch is the git/scm branch name
	Branch = "unknown"
	// SHA is the git/scm commit sha/hash
	SHA = "unknown"
	// ShortSHA is the short git/scm commit sha/hash (git rev-parse --short HEAD 7 CHARS)
	ShortSHA = "unknown"
	// Author is the git/scm entity responsible for the new commit hash (dev pushes/merges, service account fixes bad package)
	Author = "unknown"
	// BuildHost is the hostname/ci runner that build the artifact
	BuildHost = "unknown"
	// Version is the full Version, SemVer with optional supporting build metadata (1.0.0.2)
	// TODO Fix this to follow the SemVer spec here: https://semver.org/spec/v2.0.0.html#spec-item-10
	// Need to fix the artifact querier to split on '+' for build version and '-' for Prerelease
	Version = "unknown"
	// Date is the day/time (UTC) the build was created
	Date = "unknown"
	// Prerelease the name of a possible release candidate, ex: 2.1.0-rc.1, 1.0.0-alpha, 3.3.0-beta
	// note: only supply the prelease name here (rc.1, alpha, beta, etc.)
	// Prerelease is suffixed onto the Version
	// https://semver.org/spec/v2.0.0.html#spec-item-9
	Prerelease = ""
)

func FullVersion() string {
	return Version + Prerelease
}
