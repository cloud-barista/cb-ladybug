package tumblebug

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/cloud-barista/cb-mcks/src/core/app"
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
		mcisName:    mcisName,
		UserAccount: VM_USER_ACCOUNT,
	}
}

/* new instance of NLB */
func NewNLB(ns string, mcisName string, groupId string, config string) *NLBReq {
	nlb := &NLBReq{
		NLBBase: NLBBase{
			Model:  Model{Name: groupId, Namespace: ns},
			Config: config,
			Type:   "PUBLIC",
			Scope:  "REGION", Listener: NLBProtocolBase{Protocol: "TCP", Port: "6443"},
			TargetGroup: TargetGroup{NLBProtocolBase: NLBProtocolBase{Protocol: "TCP", Port: "6443"}, MCIS: mcisName, VmGroupId: groupId},
		},
		HealthChecker: HealthCheckReq{
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

/* VM */
func (self *VM) GET() (bool, error) {

	return self.execute(http.MethodGet, fmt.Sprintf("/ns/%s/mcis/%s/vm/%s", self.Namespace, self.mcisName, self.Name), nil, &self)

}

func (self *VM) POST() error {

	_, err := self.execute(http.MethodPost, fmt.Sprintf("/ns/%s/mcis/%s/vm", self.Namespace, self.mcisName), self, &self)
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
		_, err := self.execute(http.MethodDelete, fmt.Sprintf("/ns/%s/mcis/%s/vm/%s", self.Namespace, self.mcisName, self.Name), nil, app.Status{})
		if err != nil {
			return exist, err
		}
	}

	return exist, nil
}

// NLB
func (self *NLBReq) GET() (bool, error) {
	NLBRes := new(NLBRes)
	return self.execute(http.MethodGet, fmt.Sprintf("/ns/%s/mcis/%s/nlb/%s", self.Namespace, self.TargetGroup.MCIS, self.Name), nil, &NLBRes)

}

func (self *NLBReq) POST() error {
	NLBRes := new(NLBRes)
	_, err := self.execute(http.MethodPost, fmt.Sprintf("/ns/%s/mcis/%s/nlb", self.Namespace, self.TargetGroup.MCIS), self, &NLBRes)
	if err != nil {
		return err
	}

	return nil
}

func (self *NLBReq) DELETE() (bool, error) {
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
