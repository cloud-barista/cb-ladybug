# Entity

## Key
```
/ns/{namespace}/cluster/{cluster}
```

## Value
```
  {
    kind: "Cluster",
    name: "",
    status: "",
    uid: "",
    mcis: "",
    namespace: "",
    cluster-config: "",
    nodes: [
      {
        name: "",
        credential: "",
        publicIp: "",
        uid: "",
        role: "control-plane",
      },
      {
        name: "",
        credential: "",
        publicIp: "",
        uid: "",
        role: "worker",
      },
      ...
    ]
  }
```

---
## Cluster
> 클러스터 정보

|속성           |이름               |타입   |비고                                  |
|---            |---                |---    |---                                   |
|kind           |종류               |string |Cluster                               |
|name           |클러스터 명        |string |                                      |
|status         |클러스터 상태정보  |string |created/provisioning/completed/failed |
|uid            |클러스터 uid       |string |uuid                                  |
|mcis           |MCIS 명            |string |                                      |
|namespace      |MCIS 네임스페이스  |string |                                      |
|cluster-config |클러스터 연결정보  |string |Kubernetes 인 경우 kubeconfig.yaml    |


## Node
> 클러스터의 노드 정보

|속성           |이름               |타입   |비고                       |
|---            |---                |---    |---                        |
|kind           |종류               |string |Node                       |
|name           |노드명             |string |mcis vm 이름과 동일        |
|credential     |private key        |string |                           |
|publicIp       |공인 IP            |string |                           |
|uid            |노드 uid           |string |uuid                       |
|role           |역할               |string |control-plane/worker       |

