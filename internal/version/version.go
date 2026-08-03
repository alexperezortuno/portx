package version

import (
	"runtime/debug"
	"strings"
)

var (
	Version   = "dev"
	Commit    = ""
	Branch    = ""
	BuildDate = ""
	GoVersion = ""
	Platform  = ""
	Dirty     = false
)

type BuildInfo struct {
	Version   string
	Commit    string
	Branch    string
	Date      string
	GoVersion string
	Platform  string
	Dirty     bool
}

func New() *BuildInfo {
	return &BuildInfo{}
}

func Get() *BuildInfo {
	info := &BuildInfo{
		Version:   Version,
		Commit:    Commit,
		Branch:    Branch,
		Date:      BuildDate,
		GoVersion: GoVersion,
		Platform:  Platform,
		Dirty:     Dirty,
	}

	if info.GoVersion == "" {
		info.GoVersion = runtimeVersion()
	}

	return info
}

func runtimeVersion() string {
	if info, ok := debug.ReadBuildInfo(); ok {
		for _, setting := range info.Settings {
			if setting.Key == "vcs.revision" {
				if info.Main.Version != "" && !strings.Contains(info.Main.Version, "devel") {
					Version = info.Main.Version
				}
				Commit = setting.Value
			}
		}
	}
	return ""
}
