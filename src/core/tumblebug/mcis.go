package tumblebug

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/cloud-barista/cb-ladybug/src/core/app"
)

/* instance of a MCIS */
func NewMCIS(ns string, name string) *MCIS {
	return &MCIS{
		Model: Model{Name: name, Namespace: ns},
		VMs:   []VM{},
	}
}

/* instance of a VM */
func NewVM(namespace string, name string, mcisName string) *VM {
	return &VM{
		Model:       Model{Name: name, Namespace: namespace},
		McisName:    mcisName,
		UserAccount: VM_USER_ACCOUNT,
	}
}

/* new instance of NLB */
func NewNLB(ns string, mcisName string, groupId string, config string) *NLB {
	nlb := &NLB{
		NLBBase: NLBBase{
			Model:  Model{Name: groupId, Namespace: ns},
			Config: config,
			Type:   "PUBLIC",
			Scope:  "REGION", Listener: NLBProtocolBase{Protocol: "TCP", Port: "6443"},
			TargetGroup: TargetGroup{NLBProtocolBase: NLBProtocolBase{Protocol: "TCP", Port: "6443"}, MCIS: mcisName, VmGroupId: groupId},
		},
		HealthChecker: HealthCheck{
			NLBProtocolBase: NLBProtocolBase{Protocol: "TCP", Port: "22"},
			Interval:        "default", Threshold: "default", Timeout: "default",
		},
	}
	if strings.Contains(config, string(app.CSP_NCPVPC)) || strings.Contains(config, string(app.CSP_AZURE)) {
		nlb.HealthChecker.Timeout = "-1"
	}
	if strings.Contains(nlb.NLBBase.Config, string(app.CSP_GCP)) {
		nlb.HealthChecker.NLBProtocolBase.Protocol = "HTTP"
		nlb.HealthChecker.NLBProtocolBase.Port = "80"
	}

	return nlb
}

/* MCIS */
func (self *MCIS) GET() (bool, error) {

	return self.execute(http.MethodGet, fmt.Sprintf("/ns/%s/mcis/%s", self.Namespace, self.Name), nil, &self)

}

func (self *MCIS) POST() error {

	_, err := self.execute(http.MethodPost, fmt.Sprintf("/ns/%s/mcis", self.Namespace), self, &self)
	if err != nil {
		return err
	}

	return nil
}

func (self *MCIS) DELETE() (bool, error) {

	exist, err := self.GET()
	if err != nil {
		return exist, err
	}
	if exist {
		_, err := self.execute(http.MethodDelete, fmt.Sprintf("/ns/%s/mcis/%s", self.Namespace, self.Name), nil, app.Status{})
		if err != nil {
			return exist, err
		}
	}

	return exist, nil
}

func (self *MCIS) TERMINATE() error {
	_, err := self.execute(http.MethodGet, fmt.Sprintf("/ns/%s/control/mcis/%s?action=terminate", self.Namespace, self.Name), nil, app.Status{})
	if err != nil {
		return err
	}
	return nil
}

func (self *MCIS) REFINE() error {
	_, err := self.execute(http.MethodGet, fmt.Sprintf("/ns/%s/control/mcis/%s?action=refine", self.Namespace, self.Name), nil, app.Status{})
	if err != nil {
		return err
	}
	return nil
}

func (self *MCIS) FindVM(name string) *VM {
	for _, vm := range self.VMs {
		if vm.Name == name {
			return &vm
		}
	}
	return nil
}

/* VM */
func (self *VM) GET() (bool, error) {

	return self.execute(http.MethodGet, fmt.Sprintf("/ns/%s/mcis/%s/vm/%s", self.Namespace, self.McisName, self.Name), nil, &self)

}

func (self *VM) GetNameInCsp() (string, error) {
	var idsInDetail struct {
		IdInTb    string `json:"idInTb"`
		IdInSp    string `json:"idInSp"`
		IdInCsp   string `json:"idInCsp"`
		NameInCsp string `json:"nameInCsp"`
	}

	_, err := self.execute(http.MethodGet, fmt.Sprintf("/ns/%s/mcis/%s/vm/%s?option=idsInDetail", self.Namespace, self.McisName, self.Name), nil, &idsInDetail)
	if err != nil {
		return "", err
	}

	return idsInDetail.NameInCsp, nil
}

func (self *VM) POST() error {

	_, err := self.execute(http.MethodPost, fmt.Sprintf("/ns/%s/mcis/%s/vm", self.Namespace, self.McisName), self, &self)
	if err != nil {
		return err
	}

	return nil

}

func (self *VM) DELETE() (bool, error) {

	exist, err := self.GET()
	if err != nil {
		return exist, err
	}
	if exist {
		_, err := self.execute(http.MethodDelete, fmt.Sprintf("/ns/%s/mcis/%s/vm/%s", self.Namespace, self.McisName, self.Name), nil, app.Status{})
		if err != nil {
			return exist, err
		}
	}

	return exist, nil
}

// NLB
func (self *NLB) GET() (bool, error) {
	return self.execute(http.MethodGet, fmt.Sprintf("/ns/%s/mcis/%s/nlb/%s", self.Namespace, self.TargetGroup.MCIS, self.Name), nil, &self)

}

func (self *NLB) POST() error {
	_, err := self.execute(http.MethodPost, fmt.Sprintf("/ns/%s/mcis/%s/nlb", self.Namespace, self.TargetGroup.MCIS), self, &self)
	if err != nil {
		return err
	}

	return nil
}

func (self *NLB) DELETE() (bool, error) {
	exist, err := self.GET()
	if err != nil {
		return exist, err
	}
	if exist {
		_, err := self.execute(http.MethodDelete, fmt.Sprintf("/ns/%s/mcis/%s/nlb", self.Namespace, self.TargetGroup.MCIS), fmt.Sprintf(`{"connectionName" : "%s"}`, self.Config), app.Status{})
		if err != nil {
			return exist, err
		}
	}

	return exist, nil
}
