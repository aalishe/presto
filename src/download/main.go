package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.td.teradata.com/ja186051/presto/common"
)

const DOWNLOADS_DIR = "/tmp/presto-server/downloads"

const PRESTO_SERVER_VERSION = "0.152.1-t.0.4"
const PRESTO_SERVER_RPM_VERSION = "0.152.1.t.0.4-1"

const JAVA_8_VERSION = "8u111"
const JAVA_8_BUILD = "b14"

// These depends of the previous defined constants
const PRESTO_SERVER_FILE = "presto-server-rpm-%s.x86_64.rpm"                                                  // <- PRESTO_SERVER_VERSION
const PRESTO_SERVER_URL = "http://teradata-presto.s3.amazonaws.com/presto/%s/presto-server-rpm-%s.x86_64.rpm" // <- PRESTO_SERVER_VERSION, PRESTO_SERVER_VERSION

const JAVA_8_FILE = "jdk-%s-linux-x64.rpm"                                                  // <- JAVA_8_VERSION
const JAVA_8_URL = "http://download.oracle.com/otn-pub/java/jdk/%s-%s/jdk-%s-linux-x64.rpm" // <- JAVA_8_VERSION, JAVA_8_BUILD, JAVA_8_VERSION

var params map[string]string

func init() {
	params = common.Params()
	// if _, ok := params["node"]; !ok {
	// 	common.ExitWithError(errors.New("Missing parameter 'node'"))
	// }
	os.MkdirAll(DOWNLOADS_DIR, 0755)
}

// NOTE: The Java download require some headers, this could be later implented in common.Download()
func DownloadJavaOracle(path string, version string, build string) (string, error) {
	url := fmt.Sprintf(JAVA_8_URL, version, build, version)
	tmp := strings.Split(url, "/")
	filename := tmp[len(tmp)-1]
	filepath := path + "/" + filename

	// curl -s -L --header "Cookie: gpw_e24=http%3A%2F%2Fwww.oracle.com%2F; oraclelicense=accept-securebackup-cookie" ${JAVA_8_URL} -o ${DOWNLOADS_DIR}/${JAVA_8_FILE}
	cmd := exec.Command("curl", "-s", "-L", "--header", "Cookie: gpw_e24=http%3A%2F%2Fwww.oracle.com%2F; oraclelicense=accept-securebackup-cookie", url, "-o", filepath)
	_, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("Error downloading Java Oracle version %s from %s using 'curl'. Error message: %s", version, url, err)
	}
	return filepath, nil
}

func DownloadPackages() error {
	Verbose("Downloading Presto Server (if not found)\n")
	url := fmt.Sprintf(PRESTO_SERVER_URL, PRESTO_SERVER_VERSION, PRESTO_SERVER_VERSION)
	f, downloaded, err := common.DownloadIfNotExist(url, DOWNLOADS_DIR)
	if err != nil {
		return err
	}
	if downloaded {
		Verbose(fmt.Sprintf("Downloaded %s from %s\n", f, url))
	} else {
		Verbose(fmt.Sprintf("File %s was already there.\n", f))
	}

	Verbose("Downloading Java Oracle 8 (if not found)\n")
	if !common.FileExist(fmt.Sprintf(DOWNLOADS_DIR+"/"+JAVA_8_FILE, JAVA_8_VERSION)) {
		f, err := DownloadJavaOracle(DOWNLOADS_DIR, JAVA_8_VERSION, JAVA_8_BUILD)
		if err != nil {
			return err
		}
		Verbose(fmt.Sprintf("Downloaded Java Oracle %s to %s\n", JAVA_8_VERSION, f))
	} else {
		Verbose(fmt.Sprintf("Java Oracle %s was already at %s\n", JAVA_8_VERSION, fmt.Sprintf(DOWNLOADS_DIR+"/"+JAVA_8_FILE, JAVA_8_VERSION)))
	}

	return nil
}

func Verbose(message string) {
	common.Verbose(params["verbose"] == "true", message)
}

func main() {
	// fmt.Println(params)
	// imTheMaster, err := common.AmINode(params["node"])
	// if err != nil {
	// 	common.ExitWithError(err)
	// }

	si := &common.SingleInstance{
		Name: "download",
	}
	if err := si.Lock(); err != nil {
		Verbose(fmt.Sprintf("Other instance of download is running. Error: %s", err))
		fmt.Println("RUNNING")
		os.Exit(0)
	}
	defer si.Unlock()

	if err := DownloadPackages(); err != nil {
		common.ExitWithError(err)
	}
	fmt.Println("DONE")

	// imTheMaster, err := common.AmINode(params["node"])
	// if err != nil {
	// 	common.ExitWithError(err)
	// }
	// if imTheMaster {
	// 	if err := DownloadPackages(); err != nil {
	// 		common.ExitWithError(err)
	// 	}
	// 	fmt.Println("DONE")
	// } else {
	// 	fmt.Println("Event not for me, my lord")
	// }
}
