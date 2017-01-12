package common

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"syscall"
)

type SingleInstance struct {
	Name string
	file *os.File
}

func (s *SingleInstance) Filename() string {
	return filepath.Join("/var/run", fmt.Sprintf("%s.lock", s.Name))
}

func (s *SingleInstance) Lock() error {
	f, err := os.OpenFile(s.Filename(), os.O_RDWR|os.O_CREATE, 0600)
	if err != nil {
		return err
	}
	s.file = f
	flock := syscall.Flock_t{
		Type: syscall.F_WRLCK,
		Pid:  int32(os.Getpid()),
	}
	if err := syscall.FcntlFlock(s.file.Fd(), syscall.F_SETLK, &flock); err != nil {
		return fmt.Errorf("%s is already running", s.Name)
	}

	return nil
}

func (s *SingleInstance) Unlock() error {
	flock := syscall.Flock_t{
		Type: syscall.F_UNLCK,
		Pid:  int32(os.Getpid()),
	}
	if err := syscall.FcntlFlock(s.file.Fd(), syscall.F_SETLK, &flock); err != nil {
		return err
	}
	if err := s.file.Close(); err != nil {
		return err
	}
	if err := os.Remove(s.Filename()); err != nil {
		return err
	}

	return nil
}

func Payload() (string, error) {
	var payload string
	stat, _ := os.Stdin.Stat()
	if (stat.Mode() & os.ModeCharDevice) == 0 {
		bytes, err := ioutil.ReadAll(os.Stdin)
		if err != nil {
			return "", fmt.Errorf("Invalid payload input: %s\n", err)
		}
		payload = string(bytes)
	}
	return Chop(strings.ToLower(payload)), nil
}

func Params() map[string]string {
	payload, err := Payload()
	if err != nil {
		ExitWithError(err)
	}

	params := make(map[string]string)

	if strings.Contains(payload, "&") {
		for _, couple := range strings.Split(payload, "&") {
			pair := strings.Split(couple, "=")
			params[pair[0]] = pair[1]
		}
	}
	return params
}

func FileExist(filepath string) bool {
	_, err := os.Stat(filepath)
	return !os.IsNotExist(err)
}

func DownloadIfNotExist(url string, path string) (string, bool, error) {
	tmp := strings.Split(url, "/")
	filename := tmp[len(tmp)-1]

	filepath := path + "/" + filename

	if !FileExist(filepath) {
		f, err := Download(url, path)
		if err != nil {
			return f, false, fmt.Errorf("Error downloading not existing file %s from %s. Error message: %s", filepath, url, err)
		}
		return f, true, nil
	}

	return filepath, false, nil
}

func Download(url string, path string) (string, error) {
	tmp := strings.Split(url, "/")
	filename := tmp[len(tmp)-1]

	filepath := path + "/" + filename

	file, err := os.Create(filepath)
	if err != nil {
		return "", fmt.Errorf("Error creating file %s to be downloaded. Error message: %s", filepath, err)
	}
	defer file.Close()

	res, err := http.Get(url)
	if err != nil {
		return "", fmt.Errorf("Error downloading file %s from %s. Error message: %s", filepath, url, err)
	}
	defer res.Body.Close()

	_, err = io.Copy(file, res.Body)
	if err != nil {
		return "", fmt.Errorf("Error copying file %s already downloaded from %s. Error message: %s", filepath, url, err)
	}

	return filepath, nil
}

func LocalIPAddress() (string, error) {
	ifaces, err := net.Interfaces()
	if err != nil {
		return "", fmt.Errorf("Error getting network interfaces. Error message: %s", err)
	}
	for _, iface := range ifaces {
		// fmt.Printf("Checking interface: %s\n", iface.Name)
		if iface.Flags&net.FlagUp == 0 {
			continue // interface down
		}
		if iface.Flags&net.FlagLoopback != 0 {
			continue // loopback interface
		}
		addrs, err := iface.Addrs()
		if err != nil {
			return "", fmt.Errorf("Error getting addresses from interface %s. Error message: %s", iface.Name, err)
		}
		for _, addr := range addrs {
			// fmt.Printf("\tChecking IP %s\n", addr)
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}
			if ip == nil || ip.IsLoopback() {
				continue
			}
			ip = ip.To4()
			if ip == nil {
				continue // not an ipv4 address
			}
			return ip.String(), nil
		}
	}
	return "", errors.New("Local IP Address cannot be identified")
}

func Chop(str string) string {
	if len(str) <= 0 {
		return ""
	}
	return str[:len(str)-1]
}

func Verbose(verbose bool, message string) {
	if verbose {
		fmt.Print(message)
	}
}

func ExitWithError(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(0)
}

func AmINode(node string) (bool, error) {
	hostname, err := os.Hostname()
	if err != nil {
		return false, fmt.Errorf("Unable to get hostname. Error message: %s", err)
	}
	if node == hostname {
		return true, nil
	}

	localIP, err := LocalIPAddress()
	if err != nil {
		return false, err
	}
	if node == localIP {
		return true, nil
	}

	return false, nil
}
