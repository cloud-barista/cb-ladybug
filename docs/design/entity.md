# Entity

## Key
```
/ns/{namespace}/clusters/{cluster}
```

## Value
```
  {
    kind: "Cluster",
    name: "",
    status: "",
    mcis: "",
    namespace: "",
    clusterConfig: "",
    cpLeader: "",
    networkCni: "",
    nodes: [
      {
        name: "",
        credential: "",
        publicIp: "",
        role: "control-plane",
        spec: "",
        csp: "",
      },
      {
        name: "",
        credential: "",
        publicIp: "",
        role: "worker",
        spec: "",
        csp: "",
      },
      ...
    ]
  }
```

---
## Cluster
> 클러스터 정보

|속성           |이름                         |타입   |비고                                   |
|---            |---                          |---    |---                                    |
|kind           |종류                         |string |Cluster                                |
|name           |클러스터 명                  |string |                                       |
|status         |클러스터 상태정보            |string |created/provisioning/completed/failed  |
|mcis           |MCIS 명                      |string |                                       |
|namespace      |MCIS 네임스페이스            |string |                                       |
|clusterConfig  |클러스터 연결정보            |string |Kubernetes 인 경우 kubeconfig.yaml     |
|cpLeader       |control plane leader 노드명  |string |                                       |
|networkCni     |network CNI 정보             |string |                                       |


## Node
> 클러스터의 노드 정보

|속성           |이름               |타입   |비고                 |
|---            |---                |---    |---                  |
|kind           |종류               |string |Node                 |
|name           |노드명             |string |mcis vm 이름과 동일  |
|credential     |private key        |string |                     |
|publicIp       |공인 IP            |string |                     |
|role           |역할               |string |control-plane/worker |
|spec           |spec               |string |                     |
|csp            |csp 정보           |string |                     |
