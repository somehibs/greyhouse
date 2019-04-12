package version

import (
	api "git.circuitco.de/self/greyhouse/api"
)

// made this its own package because i want to share it between a bunch of places and only edit it here
// fight me

func CurrentVersion() api.NodeVersion {
	return api.NodeVersion{BuildNumber: 0,
		Version: "0.0.0.0-nonfunctional",
		UpdateNow: false,
	}
}
