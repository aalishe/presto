package main

import (
	"fmt"
	"os/exec"
	"strings"

	"github.td.teradata.com/ja186051/presto/common"
)

const DOWNLOADS_DIR = "/tmp/presto-server/downloads"

const PORT = "8080"

const PRESTO_SERVER_VERSION = "0.152.1-t.0.4"
const PRESTO_SERVER_RPM_VERSION = "0.152.1.t.0.4-1"
const JAVA_8_VERSION = "8u111"

// These depends of the previous defined constants
const PRESTO_SERVER_FILE = "presto-server-rpm-%s.x86_64.rpm" // <- PRESTO_SERVER_VERSION
const PRESTO_SERVER_RPM = "presto-server-rpm-%s.x86_64"      // <- PRESTO_SERVER_RPM_VERSION
const JAVA_8_FILE = "jdk-%s-linux-x64.rpm"                   // <- JAVA_8_VERSION

var params map[string]string

func init() {
	params = common.Params()
	// if _, ok := params["source"]; !ok {
	// 	common.ExitWithError(errors.New("Missing parameter 'Source'"))
	// }
}

func IsPrestoInstalled() (bool, error) {
	cmd := exec.Command("rpm", "-qa")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return false, fmt.Errorf("Error checking if Presto is installed using 'rpm'. Error message: %s", err)
	}
	// fmt.Println(string(out))
	return strings.Contains(string(out), fmt.Sprintf(PRESTO_SERVER_RPM, PRESTO_SERVER_RPM_VERSION)), nil
}

func InstallPresto(path string) error {
	installed, err := IsPrestoInstalled()
	if err != nil {
		return err
	}
	if !installed {
		javafilepath := fmt.Sprintf(path+"/"+JAVA_8_FILE, JAVA_8_VERSION)
		prestofilepath := fmt.Sprintf(path+"/"+PRESTO_SERVER_FILE, PRESTO_SERVER_VERSION)

		cmd := exec.Command("rpm", "-i", prestofilepath, javafilepath)
		_, err := cmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("Error installing Presto and Java Oracle using 'rpm'. Error message: %s", err)
		}
		Verbose("Presto Server successfully installed\n")
	} else {
		Verbose("Presto already installed\n")
	}

	return nil
}

func download(item string, filepath string) error {
	url := fmt.Sprintf("http://%s:%s/%s", params["source"], PORT, item)
	if !common.FileExist(filepath) {
		Verbose(fmt.Sprintf("Downloading (%s) from %s\n", filepath, url))
		_, err := common.Download(url, DOWNLOADS_DIR)
		if err != nil {
			return err
		}
	} else {
		Verbose(fmt.Sprintf("File (%s) was already there.\n", filepath))
	}
	return nil
}

func Verbose(message string) {
	common.Verbose(params["verbose"] == "true", message)
}

func main() {
	installed, err := IsPrestoInstalled()
	if err != nil {
		common.ExitWithError(err)
	}
	if installed {
		Verbose("Presto Server and Java Oracle 8 already installed\n")
		fmt.Println("DONE")
	} else {
		javafilepath := fmt.Sprintf(DOWNLOADS_DIR+"/"+JAVA_8_FILE, JAVA_8_VERSION)
		prestofilepath := fmt.Sprintf(DOWNLOADS_DIR+"/"+PRESTO_SERVER_FILE, PRESTO_SERVER_VERSION)

		if !common.FileExist(javafilepath) {
			common.ExitWithError(fmt.Errorf("Error: Java Oracle 8 rpm not found (%s)", javafilepath))
		}
		Verbose(fmt.Sprintf("Java Oracle 8 rpm found at %s\n", javafilepath))
		if !common.FileExist(prestofilepath) {
			common.ExitWithError(fmt.Errorf("Error: Presto Server rpm not found (%s)", prestofilepath))
		}
		Verbose(fmt.Sprintf("Presto Server rpm found at %s\n", prestofilepath))

		if err := InstallPresto(DOWNLOADS_DIR); err != nil {
			common.ExitWithError(err)
		} else {
			fmt.Println("DONE")
		}

		// imTheSource, err := common.AmINode(params["source"])
		// if err != nil {
		// 	common.ExitWithError(err)
		// }
		// if !imTheSource {
		// 	Verbose(fmt.Sprintf("Downloading Presto Server from http://%s:%s/presto\n", params["source"], PORT))
		// 	if err = download("presto", prestofilepath); err != nil {
		// 		common.ExitWithError(err)
		// 	}
		// 	Verbose(fmt.Sprintf("Downloading Java Oracle 8 from http://%s:%s/java\n", params["source"], PORT))
		// 	if err = download("java", javafilepath); err != nil {
		// 		common.ExitWithError(err)
		// 	}
		// } else {
		// 	if !common.FileExist(javafilepath) {
		// 		common.ExitWithError(fmt.Errorf("Error: Java Oracle 8 rpm not found (%s)", javafilepath))
		// 	}
		// 	Verbose(fmt.Sprintf("Java Oracle 8 rpm found at %s\n", javafilepath))
		// 	if !common.FileExist(prestofilepath) {
		// 		common.ExitWithError(fmt.Errorf("Error: Presto Server rpm not found (%s)", prestofilepath))
		// 	}
		// 	Verbose(fmt.Sprintf("Presto Server rpm found at %s\n", prestofilepath))
		// }
		// if err := InstallPresto(DOWNLOADS_DIR); err != nil {
		// 	common.ExitWithError(err)
		// } else {
		// 	fmt.Println("DONE")
		// }
	}
}
