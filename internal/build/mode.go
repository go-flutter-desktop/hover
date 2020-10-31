package build

type Mode struct {
	Name  string
	IsAot bool
}

var DebugMode = Mode{
	Name:  "debug_unopt",
	IsAot: false,
}

// JitReleaseMode is the same debug build, but disables the terminal windows on Windows
var JitReleaseMode = Mode{
	Name:  "debug_unopt",
	IsAot: false,
}

var ReleaseMode = Mode{
	Name:  "release",
	IsAot: true,
}

var ProfileMode = Mode{
	Name:  "profile",
	IsAot: true,
}
