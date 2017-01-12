package main

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"

	"github.td.teradata.com/ja186051/presto/common"
)

const DOWNLOADS_DIR = "/tmp/presto-server/downloads"

const PORT = "8080"

const PRESTO_SERVER_VERSION = "0.152.1-t.0.4"
const PRESTO_SERVER_RPM_VERSION = "0.152.1.t.0.4-1"

const JAVA_8_APP_VERSION = "1.8.0_111"
const JAVA_8_VERSION = "8u111"
const JAVA_8_BUILD = "b14"

// These depends of the previous defined constants
const PRESTO_SERVER_FILE = "presto-server-rpm-%s.x86_64.rpm" // <- PRESTO_SERVER_VERSION
const JAVA_8_FILE = "jdk-%s-linux-x64.rpm"                   // <- JAVA_8_VERSION

var params map[string]string

func init() {
	params = common.Params()
	if _, ok := params["node"]; !ok {
		common.ExitWithError(errors.New("Missing parameter 'node'"))
	}
}

func HandleFileRequest(w http.ResponseWriter, r *http.Request) {
	var filename string
	url := r.URL.Path

	if url == "/java" {
		filename = fmt.Sprintf(JAVA_8_FILE, JAVA_8_VERSION)
	} else if url == "/presto" {
		filename = fmt.Sprintf(PRESTO_SERVER_FILE, PRESTO_SERVER_VERSION)
	} else {
		http.Error(w, fmt.Sprintf("URL path %s is not valid", url), 400)
		return
	}
	filepath := DOWNLOADS_DIR + "/" + filename
	Verbose(fmt.Sprintf("Trasfer of file %s\n", filename))
	file, err := os.Open(filepath)
	if err != nil {
		http.Error(w, fmt.Sprintf("File %s not found", filename), 400)
		return
	}
	defer file.Close()

	fileHeader := make([]byte, 512)
	file.Read(fileHeader)
	fileContentType := http.DetectContentType(fileHeader)

	fileStat, _ := file.Stat()
	fileSize := strconv.FormatInt(fileStat.Size(), 10)

	w.Header().Set("Content-Disposition", "attachment; filename="+filename)
	w.Header().Set("Content-Type", fileContentType)
	w.Header().Set("Content-Length", fileSize)

	file.Seek(0, 0)
	io.Copy(w, file)
	return
}

func SharePackages() error {
	Verbose(fmt.Sprintf("Sharing at http://%s:%s/\n", params["node"], PORT))

	http.HandleFunc("/", HandleFileRequest)
	err := http.ListenAndServe(":"+PORT, nil)
	if err != nil {
		return fmt.Errorf("Failed to create the server on port %s. Error message: %s", PORT, err)
	}

	return nil
}

func Verbose(message string) {
	common.Verbose(params["verbose"] == "true", message)
}

func main() {
	// fmt.Println(params)
	imTheMaster, err := common.AmINode(params["node"])
	if err != nil {
		common.ExitWithError(err)
	}
	if imTheMaster {
		if err := SharePackages(); err != nil {
			common.ExitWithError(err)
		}
	}
}
