package cmd

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

func init() {
	pluginCmd.AddCommand(pluginListCmd)
	rootCmd.AddCommand(pluginCmd)
}

var pluginCmd = &cobra.Command{
	Use:   "plugin",
	Short: "Tools for plugins",
	Long:  "A collection of commands to help with finding desktop implementations for existing plugins.",
}

type pubDependency struct {
	name       string
	version    string
	transitive bool

	android bool
	ios     bool
	desktop bool
}

func (p pubDependency) isPlugin() bool {
	return p.android || p.ios || p.desktop
}

func (p pubDependency) platforms() []string {
	var platforms []string
	if p.android {
		platforms = append(platforms, "android")
	}
	if p.ios {
		platforms = append(platforms, "ios")
	}
	if p.desktop {
		platforms = append(platforms, "desktop")
	}
	return platforms
}

func listDependencies() ([]pubDependency, error) {
	listCmd := exec.Command(flutterBin, "pub", "deps", "--", "--style=compact", "--no-dev")
	listOutput, err := listCmd.Output()
	if err != nil {
		return nil, errors.Wrap(err, "failed to list dependencies for project")
	}

	pubcachePath, err := findPubcachePath()
	if err != nil {
		return nil, errors.Wrap(err, "failed to find path for pub-cache")
	}

	var list []pubDependency
	listReader := bufio.NewReader(bytes.NewBuffer(listOutput))
	var inDependencies bool
	var inTransitive bool
	for {
		line, err := listReader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				return list, nil
			}
			return nil, errors.Wrap(err, "failed reading dependency list")
		}
		line = strings.TrimSpace(line)

		if !inDependencies {
			if line == "dependencies:" {
				inDependencies = true
			}
			continue
		}
		if !inTransitive {
			if line == "transitive dependencies:" {
				inTransitive = true
				continue
			}
		}

		if !strings.HasPrefix(line, "- ") {
			continue
		}
		lineParts := strings.SplitN(line[2:], " ", 3)
		if len(lineParts) != 2 && len(lineParts) != 3 {
			return nil, errors.Errorf("failed to parse dependency line `%s`", line)
		}

		p := pubDependency{
			name:       lineParts[0],
			version:    lineParts[1],
			transitive: inTransitive,
		}

		detectPlatform := func(platform string) (bool, error) {
			platformPath := filepath.Join(pubcachePath, "hosted", "pub.dartlang.org", p.name+"-"+p.version, platform)
			stat, err := os.Stat(platformPath)
			if err != nil {
				if os.IsNotExist(err) {
					return false, nil
				}
				return false, errors.Wrapf(err, "failed to stat %s", platformPath)
			}
			return stat.IsDir(), nil
		}
		p.android, err = detectPlatform("android")
		if err != nil {
			return nil, err
		}
		p.ios, err = detectPlatform("ios")
		if err != nil {
			return nil, err
		}
		p.desktop, err = detectPlatform("desktop")
		if err != nil {
			return nil, err
		}

		list = append(list, p)
	}
}

var pluginListCmd = &cobra.Command{
	Use:   "list",
	Short: "List plugins in the application",
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) > 0 {
			return errors.New("does not take arguments")
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		assertInFlutterProject()
		dependencyList, err := listDependencies()
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}

		var hasPlugins bool
		for _, dep := range dependencyList {
			if !dep.isPlugin() {
				continue
			}

			if hasPlugins {
				fmt.Println("")
			}
			hasPlugins = true

			fmt.Println(dep.name)
			fmt.Printf("  version:   %s\n", dep.version)
			fmt.Printf("  platforms: [%s]\n", strings.Join(dep.platforms(), ", "))
		}
	},
}
