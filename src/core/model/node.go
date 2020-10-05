package model

type Node struct {
	Model
	Credential string `json:"credential"`
	PublicIP   string `json:"publicIp"`
}

func NewNode() *Node {
	return &Node{
		Model: Model{Kind: KIND_NODE},
	}
}
