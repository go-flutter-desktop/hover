package androidmanifest

import (
	"encoding/xml"
	"io/ioutil"
	"os"
	"strings"

	"github.com/go-flutter-desktop/hover/internal/log"
)

// AndroidManifest is a file that describes the essential information about
// an android app.
type AndroidManifest struct {
	Package string `xml:"package,attr"`
}

// AndroidOrganizationName fetch the android package name (default:
// 'com.example').
// Can by set upon flutter create (--org flag)
//
// If errors occurs when reading the android package name, the string value
// will correspond to 'hover.failed.to.retrieve.package.name'
func AndroidOrganizationName() string {
	// Default value
	androidManifestFile := "android/app/src/main/AndroidManifest.xml"

	// Open AndroidManifest file
	xmlFile, err := os.Open(androidManifestFile)
	if err != nil {
		log.Errorf("Failed to retrieve the organization name: %v", err)
		return "hover.failed.to.retrieve.package.name"
	}
	defer xmlFile.Close()

	byteXMLValue, err := ioutil.ReadAll(xmlFile)
	if err != nil {
		log.Errorf("Failed to retrieve the organization name: %v", err)
		return "hover.failed.to.retrieve.package.name"
	}

	var androidManifest AndroidManifest
	err = xml.Unmarshal(byteXMLValue, &androidManifest)
	if err != nil {
		log.Errorf("Failed to retrieve the organization name: %v", err)
		return "hover.failed.to.retrieve.package.name"
	}
	javaPackage := strings.Split(androidManifest.Package, ".")
	orgName := strings.Join(javaPackage[:len(javaPackage)-1], ".")
	if orgName == "" {
		return "hover.failed.to.retrieve.package.name"
	}
	return orgName
}
