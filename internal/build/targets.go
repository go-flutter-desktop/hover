package build

import (
	"errors"
	"fmt"
	"strings"
)

type TargetPlatform = string
type TargetPackagingFormat = string

type Target struct {
	Platform        TargetPlatform
	PackagingFormat TargetPackagingFormat
}

type buildTargetPlatforms struct {
	Linux   TargetPlatform
	Darwin  TargetPlatform
	Windows TargetPlatform
	All     TargetPlatform
}

type buildTargetPackagingFormats struct {
	AppImage TargetPackagingFormat
	Deb      TargetPackagingFormat
	Snap     TargetPackagingFormat
	Bundle   TargetPackagingFormat
	Pkg      TargetPackagingFormat
	Msi      TargetPackagingFormat
	All      TargetPackagingFormat
}

var TargetPlatforms = &buildTargetPlatforms{
	"linux",
	"darwin",
	"windows",
	"*",
}

var TargetPackagingFormats = &buildTargetPackagingFormats{
	"appimage",
	"deb",
	"snap",
	"bundle",
	"pkg",
	"msi",
	"*",
}

var linuxTarget = Target{
	Platform:        TargetPlatforms.Linux,
	PackagingFormat: "",
}
var darwinTarget = Target{
	Platform:        TargetPlatforms.Darwin,
	PackagingFormat: "",
}
var windowsTarget = Target{
	Platform:        TargetPlatforms.Windows,
	PackagingFormat: "",
}
var PlatformTargets = []Target{
	linuxTarget,
	darwinTarget,
	windowsTarget,
}
var linuxPackagingTargets = []Target{
	{
		Platform:        TargetPlatforms.Linux,
		PackagingFormat: TargetPackagingFormats.AppImage,
	},
	{
		Platform:        TargetPlatforms.Linux,
		PackagingFormat: TargetPackagingFormats.Deb,
	},
	{
		Platform:        TargetPlatforms.Linux,
		PackagingFormat: TargetPackagingFormats.Snap,
	},
}
var darwinPackagingTargets = []Target{
	{
		Platform:        TargetPlatforms.Darwin,
		PackagingFormat: TargetPackagingFormats.Bundle,
	},
	{
		Platform:        TargetPlatforms.Darwin,
		PackagingFormat: TargetPackagingFormats.Pkg,
	},
}
var windowsPackagingTargets = []Target{
	{
		Platform:        TargetPlatforms.Windows,
		PackagingFormat: TargetPackagingFormats.Msi,
	},
}
var PackagingTargets = append(append(append([]Target{}, linuxPackagingTargets...), darwinPackagingTargets...), windowsPackagingTargets..., )

func AreValidBuildTargets(parameter string, requirePackagingFormat bool) error {
	targets, err := ParseBuildTargets(parameter, requirePackagingFormat)
	if requirePackagingFormat {
		for _, target := range targets {
			if target.PackagingFormat == "" {
				return errors.New(fmt.Sprintf("build target has no packaging format: %s", target.Platform))
			}
		}
	}
	return err
}

func ParseBuildTargets(parameter string, requirePackagingFormat bool) ([]Target, error) {
	targetStrings := strings.Split(parameter, ",")
	var targets []Target
	for _, targetString := range targetStrings {
		parts := strings.Split(targetString, "-")
		target := Target{}
		if len(parts) == 0 || len(parts) > 2 {
			return nil, errors.New(fmt.Sprintf("not a valid target: `%s`", targetString))
		}
		if len(parts) > 0 {
			target.Platform = parts[0]
		}
		if len(parts) > 1 {
			target.PackagingFormat = parts[1]
		} else {
			target.PackagingFormat = ""
		}
		switch target.Platform {
		case TargetPlatforms.Linux:
		case TargetPlatforms.Darwin:
		case TargetPlatforms.Windows:
		case TargetPlatforms.All:
		default:
			return nil, errors.New(fmt.Sprintf("not a valid target platform: `%s`", target.Platform))
		}
		switch target.PackagingFormat {
		case TargetPackagingFormats.AppImage:
		case TargetPackagingFormats.Deb:
		case TargetPackagingFormats.Snap:
		case TargetPackagingFormats.Bundle:
		case TargetPackagingFormats.Pkg:
		case TargetPackagingFormats.Msi:
		case TargetPackagingFormats.All:
		case "":
		default:
			return nil, errors.New(fmt.Sprintf("not a valid target packaging format: `%s`", target.PackagingFormat))
		}
		if target.Platform == TargetPlatforms.Linux && (target.PackagingFormat == TargetPackagingFormats.All || target.PackagingFormat == TargetPackagingFormats.AppImage || target.PackagingFormat == TargetPackagingFormats.Deb || target.PackagingFormat == TargetPackagingFormats.Snap) {
			// Is a valid linux packaging combination
		} else if target.Platform == TargetPlatforms.Darwin && (target.PackagingFormat == TargetPackagingFormats.All || target.PackagingFormat == TargetPackagingFormats.Bundle || target.PackagingFormat == TargetPackagingFormats.Pkg) {
			// Is a valid darwin packaging combination
		} else if target.Platform == TargetPlatforms.Windows && (target.PackagingFormat == TargetPackagingFormats.All || target.PackagingFormat == TargetPackagingFormats.Msi) {
			// Is a valid windows packaging combination
		} else if target.PackagingFormat == "" {
			// Has no packaging format
		} else if target.Platform == TargetPlatforms.All && target.PackagingFormat == TargetPackagingFormats.All {
			// Is a valid all platform / all packaging combination
		} else {
			return nil, errors.New(fmt.Sprintf("not a valid target platform / packaging format combination: `%s-%s`", target.Platform, target.PackagingFormat))
		}

		/// Set darwin-bundle as a required packaging format for darwin-pkg
		if target.Platform == TargetPlatforms.Darwin && target.PackagingFormat == TargetPackagingFormats.Pkg {
			targets = append(targets, Target{
				Platform:        TargetPlatforms.Darwin,
				PackagingFormat: TargetPackagingFormats.Bundle,
			})
		}
		if target.Platform == TargetPlatforms.All && target.PackagingFormat == "" && !requirePackagingFormat {
			targets = append(targets, PlatformTargets...)
			break
		}
		if target.Platform == TargetPlatforms.Linux && target.PackagingFormat == TargetPackagingFormats.All {
			targets = append(targets, linuxPackagingTargets...)
			break
		}
		if target.Platform == TargetPlatforms.Darwin && target.PackagingFormat == TargetPackagingFormats.All {
			targets = append(targets, darwinPackagingTargets...)
			break
		}
		if target.Platform == TargetPlatforms.Windows && target.PackagingFormat == TargetPackagingFormats.All {
			targets = append(targets, windowsPackagingTargets...)
			break
		}
		if target.Platform == TargetPlatforms.All && target.PackagingFormat == TargetPackagingFormats.All || (target.Platform == TargetPlatforms.All && requirePackagingFormat) {
			targets = append(targets, PackagingTargets...)
			break
		}
		targets = append(targets, target)
	}
	return uniqueTargets(targets), nil
}

func uniqueTargets(intSlice []Target) []Target {
	keys := make(map[Target]bool)
	var list []Target
	for _, entry := range intSlice {
		if _, value := keys[entry]; !value {
			keys[entry] = true
			list = append(list, entry)
		}
	}
	return list
}
