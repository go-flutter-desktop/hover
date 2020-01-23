package build

type Mode struct {
	Name  string
	IsAot bool
}

var DebugMode = Mode{
	Name:  "debug",
	IsAot: false,
}

var DebugReleaseMode = Mode{
	Name:  "debug-release",
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
