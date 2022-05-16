# Hover - Run Flutter apps on the desktop with hot-reload

Hover is a simple build tool to create [Flutter](https://flutter.dev) desktop applications.

**Hover is brand new and under development, it should be considered alpha. Anything can break, please backup your data before using hover**

Hover is part of the [go-flutter](https://github.com/go-flutter-desktop/go-flutter) project. Please report issues at the [go-flutter issue tracker](https://github.com/go-flutter-desktop/go-flutter/issues/).

## Install

Hover uses [Go](https://golang.org) to build your Flutter application to desktop. Hover itself is also written using the Go language. You will need to [install go](https://golang.org/doc/install) on your development machine.

Run `go version` and make sure that your Go version is 1.13 or higher.

Then install hover by running this in your home directory:

```bash
GO111MODULE=on go get -u -a github.com/go-flutter-desktop/hover@latest
```
Or windows:
```
set GO111MODULE=on
go install github.com/go-flutter-desktop/hover@latest
```
Or for powershell:
```powershell
$env:GO111MODULE="on"; go get -u -a github.com/go-flutter-desktop/hover@latest
```
Make sure the hover binary is on your `PATH` (defaults are `$GOPATH/bin` or `$HOME/go/bin`)

Run the same command to update when a newer version becomes available.

Install these dependencies:

* You need to make sure you have a C compiler.  
  The recommended C compiler are documented [here](https://github.com/golang/go/wiki/InstallFromSource#install-c-tools).

* You need to make sure you have dependencies of GLFW:
  * On macOS, you need Xcode or Command Line Tools for Xcode (`xcode-select --install`) for required headers and libraries.
  * On Ubuntu/Debian-like Linux distributions, you need `libgl1-mesa-dev xorg-dev` packages.
  * On CentOS/Fedora-like Linux distributions, you need `libX11-devel libXcursor-devel libXrandr-devel libXinerama-devel mesa-libGL-devel libXi-devel` packages.
  * See [here](http://www.glfw.org/docs/latest/compile.html#compile_deps) for full details.

## Getting started with an existing Flutter project

This assumes you have an existing flutter project which you want to run on desktop. If you don't have a project yet, follow the flutter tutorial for setting up a new project first.

### Init project for hover

cd into a flutter project.

```bash
cd projects/simpleApplication
```

The first time you use hover for a project, you'll need to initialize the project for use with hover. An argument can be passed to `hover init` to set the project path. This is usually the path for your project on github or a self-hosted git service. _If you are unsure use `hover init` without a path. You can change the path later._

```bash
hover init github.com/my-organization/simpleApplication
```

This creates the directory `go` and adds boilerplate files such as Go code and a default logo.

Optionally, you may add [plugins](https://github.com/go-flutter-desktop/plugins) to `go/cmd/options.go`  
Optionally, change the logo in `go/assets/logo.png`, which is used as icon for the window.

### Run with hot-reload

To run the application and attach flutter for hot-reload support:

```bash
hover run
```

The hot-reload is manual because you'll need to press 'r' in the terminal to hot-reload the application.

By default, hover uses the file `lib/main_desktop.dart` as entrypoint. You may specify a different endpoint by using the `--target` flag.

#### IDE integration

##### VSCode

Please try the [experimental Hover extension for VSCode](https://marketplace.visualstudio.com/items?itemName=go-flutter.hover).

If you want to manually integrate with VSCode, read this [issue](https://github.com/go-flutter-desktop/go-flutter/issues/129#issuecomment-513590141).

##### Emacs

Check [hover.el](https://github.com/ericdallo/hover.el) packge for emacs integration.

### Build standalone application

To create a standalone release (JIT mode) build run this command:

```bash
hover build linux # or darwin or windows
```

You can create a build for any of the supported OSs using cross-compiling which needs [Docker to be installed](https://docs.docker.com/install/).
Then run the command from above and it will do everything for you.

The output will be in `go/build/outputs/linux` or windows or darwin.

To start the binary: (replace `yourApplicationName` with your app name)

```bash
./go/build/outputs/linux/yourApplicationName
```

It's possible to zip the whole dir `go/build/outputs/linux` and ship it to a different machine.

### Packaging

You can package your application for different packaging formats.  
First initialize the packaging format:

```bash
hover init-packaging linux-appimage
```

Update the configuration files located in `go/packaging/linux-appimage/`to your needs.  
Then create a build and package it using this command:

```bash
hover build linux-appimage
```

The packaging output is placed in `go/build/outputs/linux-appimage/`

To get a list of all available packaging formats run:

```bash
hover build --help
```

### Flavors

Hover supports different application flavors via `--flavor MY_FLAVOR` command.
If you wish to create a new flavor for you application, 
simply copy `go/hover.yaml` into `go/hover-MY_FLAVOR.yaml` and modify contents as needed.
If no flavor is specified, Hover will always default to `hover.yaml`

```
hover run --flavor develop || hover build --flavor develop
// hover-develop.yaml
```


## Issues

Please report issues at the [go-flutter issue tracker](https://github.com/go-flutter-desktop/go-flutter/issues/).
