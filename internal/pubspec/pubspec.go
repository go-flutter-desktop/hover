package pubspec

import (
	"fmt"
	"gopkg.in/yaml.v2"
	"os"
)

type PubSpec struct {
	Name         string
	Description  string
	Version      string
	Author       string
	Dependencies map[string]interface{}
}

var pubspec = PubSpec{}

func GetPubSpec() PubSpec {
	{
		if pubspec.Name == "" {
			file, err := os.Open("pubspec.yaml")
			if err != nil {
				if os.IsNotExist(err) {
					fmt.Println("hover: Error: No pubspec.yaml file found.")
					goto Fail
				}
				fmt.Printf("hover: Failed to open pubspec.yaml: %v\n", err)
				os.Exit(1)
			}
			defer file.Close()

			err = yaml.NewDecoder(file).Decode(&pubspec)
			if err != nil {
				fmt.Printf("hover: Failed to decode pubspec.yaml: %v\n", err)
				goto Fail
			}
			if _, exists := pubspec.Dependencies["flutter"]; !exists {
				fmt.Println("hover: Missing `flutter` in pubspec.yaml dependencies list.")
				goto Fail
			}
		}

		return pubspec
	}

Fail:
	fmt.Println("hover: This command should be run from the root of your Flutter project.")
	os.Exit(1)
	return PubSpec{}
}
