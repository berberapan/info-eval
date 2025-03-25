package vcs

import "runtime/debug"

var readBuildInfo = func() (*debug.BuildInfo, bool) {
	return debug.ReadBuildInfo()
}

func Version() string {
	info, ok := readBuildInfo()
	if ok {
		return info.Main.Version
	}
	return ""
}
