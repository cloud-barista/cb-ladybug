package lang

import (
	"fmt"
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
