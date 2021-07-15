package util

import (
	"github.com/Masterminds/semver"
)

var (
	// The version MUST be "v"+semanticVersion, it should be one of the following options
	// * "vX.Y.Z" for released version, e.g. "v1.2.3"
	// * "v0.0.0-X-Y" for unreleased versions, e.g.
	//   - "v0.0.0-latest-0" ("latest" version, i.e. latest build on the "main" branch or local dev build)
	version = "v0.0.0-latest-0"
	logger  = NewLogger()
	Version semver.Version
)

func init() {
	logger.Info("version", "version", version)
	Version = *semver.MustParse(version)
}
