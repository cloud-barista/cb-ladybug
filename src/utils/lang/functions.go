package lang

import (
	"fmt"
	"os"
	"regexp"

	"github.com/google/uuid"
)

// NVL is null value logic
func NVL(str string, def string) string {
	if len(str) == 0 {
		return def
	}
	return str
}

// get store key
func GetStoreKey(namespace string, clusterName string) string {
	return fmt.Sprintf("/ns/%s/cluster/%s", namespace, clusterName)
}

// get scripts file path
func GetScriptsPath() string {
	var sourcePath string
	pwd, err := os.Getwd()
	if err != nil {
		sourcePath = os.Getenv("GOPATH") + "/src/github.com/cloud-barista/cb-ladybug/src/scripts/"
	} else {
		pathRegex, _ := regexp.Compile("(.*?)\\/github.com\\/cloud-barista\\/cb-ladybug(.*?)")
		if pathRegex.MatchString(pwd) {
			res := pathRegex.FindStringSubmatch(pwd)
			pwd = res[0]
		}
		sourcePath = pwd + "/src/scripts/"
	}
	return sourcePath
}

// for worker node join command
func GetWorkerJoinCmd(cpInitResult string) string {
	var join1, join2 string
	joinRegex, _ := regexp.Compile("kubeadm\\sjoin\\s(.*?)\\s--token\\s(.*?)\\s")
	joinRegex2, _ := regexp.Compile("--discovery-token-ca-cert-hash\\ssha256:(.*?)\\n")

	if joinRegex.MatchString(cpInitResult) {
		res := joinRegex.FindStringSubmatch(cpInitResult)
		join1 = res[0]
	}
	if joinRegex2.MatchString(cpInitResult) {
		res := joinRegex2.FindStringSubmatch(cpInitResult)
		join2 = res[0]
	}

	return fmt.Sprintf("sudo %s %s", join1, join2)
}

// get uuid
func GetUid() string {
	return uuid.New().String()
}
