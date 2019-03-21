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

### Use

cd into a flutter project.

```bash
cd to/my/project
```

Then, only the first time, you'll need to initialize for desktop.

```bash
hover init
```

This creates the directory `desktop` and adds Go files.

You may add plugins in `desktop/cmd/options.go`

Optionally change the logo in `desktop/assets/logo.png`, which is used as icon for the window.

To build and execute, run:

```bash
hover build
cd desktop/build/outputs/linux
./yourApplicationName
```
