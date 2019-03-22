# hover - Release Flutter applications on the desktop

`hover` is a simple tool to help with building a flutter application for desktop.

**Hover is brand new and under development, it should be considered alpha. Anything can break, please backup your data before using hover**

## Getting started

### Install

First install go, then install hover like this:

```bash
go get -u github.com/go-flutter-desktop/hover
```

Run the same command to update when a newer version becomes available.

Install these dependencies:

* You need to make sure you have dependencies of GLFW:
	* On macOS, you need Xcode or Command Line Tools for Xcode (`xcode-select --install`) for required headers and libraries.
	* On Ubuntu/Debian-like Linux distributions, you need `libgl1-mesa-dev xorg-dev` packages.
	* On CentOS/Fedora-like Linux distributions, you need `libX11-devel libXcursor-devel libXrandr-devel libXinerama-devel mesa-libGL-devel libXi-devel` packages.
	* See [here](http://www.glfw.org/docs/latest/compile.html#compile_deps) for full details.

### Use

cd into a flutter project.

```bash
cd projects/simpleApplication
```

The first time you use hover for a project, you'll need to initialize the project for desktop. `hover init` requires a project path. This is usualy the path for your project on github or a self-hosted git service. _If you are unsure, just make something up, it can always be changed later._

```bash
hover init github.com/my-organization/simpleApplication
```

This creates the directory `desktop` and adds boilerplate files such as Go code and a default logo.

Optionally, you may add [plugins](https://github.com/go-flutter-desktop/plugins) to `desktop/cmd/options.go`

Optionally, change the logo in `desktop/assets/logo.png`, which is used as icon for the window.

To build and execute, run:

```bash
hover build
./desktop/build/outputs/linux/yourApplicationName
```

## Issues

Please report issues at the [go-flutter issue tracker](https://github.com/go-flutter-desktop/go-flutter/issues/).
