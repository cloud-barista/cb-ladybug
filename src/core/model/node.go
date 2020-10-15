package model

type Node struct {
	Name       string `json:"name"`
	Credential string `json:"Credential"`
	PublicIP   string `json:"publicIP"`
	UId        string `json:"uId"`
	Role       string `json:"role"`
}

func NewNode(vm VM) *Node {
	return &Node{
		Name:       vm.Name,
		Credential: vm.Credential,
		PublicIP:   vm.PublicIP,
		UId:        vm.UId,
		Role:       vm.Role,
	}
}
