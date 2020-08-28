// This package allows us to inject build time variables such as commit hash
// When we run this in Docker in Kubernetes, we want to be able to
// introspectively identify what source code generated the runtime we
// encounter.
// See https://github.com/StevenACoffman/small for more information
// Also https://blog.alexellis.io/inject-build-time-vars-golang/
// These variables will be injected at build time in the Docker container
// Please see the README.md for more information as to how this is done here.
package version

import (
	"fmt"
)

// 	Populated during build
var (
	AppName      = "unknown" // What the app is called
	Project      = "unknown" // Which GCP Project e.g. khan-internal-services
	Date         = ""        // Build Date in RFC3339 e.g. $(date -u +"%Y-%m-%dT%H:%M:%SZ")
	GitCommit    = "?"       // Git Commit sha of source
	Version      = "v0.0.0"  // go mod version: v0.0.0-20200214070026-92e9ce6ff79f
	HumanVersion = fmt.Sprintf(
		"%s %s %s (%s) on %s",
		AppName,
		Project,
		Version,
		GitCommit,
		Date,
	)
)
