# Entity

<<<<<<< HEAD
<<<<<<< HEAD
=======
>>>>>>> Update entity
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
<<<<<<< HEAD
<<<<<<< HEAD
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
=======
    uId: "",
=======
    uid: "",
>>>>>>> Fix json in model.Cluster, model.Node
    mcis: "",
    namespace: "",
    cluster-config: "",
    nodes: [
      {
        name: "",
<<<<<<< HEAD
        Credential: "",
        publicIP: "",
        uId: "",
>>>>>>> Update entity
=======
        credential: "",
        publicIp: "",
        uid: "",
>>>>>>> Fix json in model.Cluster, model.Node
        role: "control-plane",
      },
      {
        name: "",
<<<<<<< HEAD
<<<<<<< HEAD
        credential: "",
        publicIp: "",
        uid: "",
=======
        Credential: "",
        publicIP: "",
        uId: "",
>>>>>>> Update entity
=======
        credential: "",
        publicIp: "",
        uid: "",
>>>>>>> Fix json in model.Cluster, model.Node
        role: "worker",
      },
      ...
    ]
  }
```

---
<<<<<<< HEAD
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
=======
## Cluster
> 클러스터 정보

|속성           |이름               |타입   |KEY|비고                               |
|---            |---                |---    |---|---                                |
|name           |클러스터 명        |string |NN |                                   |
|uid            |클러스터 uid       |string |PK |입력시 값 생성                     |
|namespace      |MCIS 네임스페이스  |string |NN |                                   |
|mcis           |MCIS 명            |string |NN |                                   |
|cluster-config |클러스터 연결정보  |string |   |Kubernetes 인 경우 kubeconfig.yaml |
=======
## Cluster
> 클러스터 정보

|속성           |이름               |타입   |비고                               |
|---            |---                |---    |---                                |
|kind           |종류        |string |Cluster                                   |
|name           |클러스터 명        |string |                                   |
|status            |클러스터 상태정보       |string |created/provisioning/completed/failed|
|uid            |클러스터 uid       |string |uuid                     |
|mcis           |MCIS 명            |string |                                   |
|namespace      |MCIS 네임스페이스  |string |                                   |
<<<<<<< HEAD
|clusterConfig |클러스터 연결정보  |string |Kubernetes 인 경우 kubeconfig.yaml |
>>>>>>> Update entity

>>>>>>> Add base source code
=======
|cluster-config |클러스터 연결정보  |string |Kubernetes 인 경우 kubeconfig.yaml |
>>>>>>> Fix json in model.Cluster, model.Node


## Node
> 클러스터의 노드 정보

<<<<<<< HEAD
<<<<<<< HEAD
<<<<<<< HEAD
|속성           |이름               |타입   |비고                       |
|---            |---                |---    |---                        |
|name           |노드명             |string |mcis vm 이름과 동일        |
|credential     |private key        |string |                           |
|publicIp       |공인 IP            |string |                           |
|uid            |노드 uid           |string |uuid                       |
|role           |역할               |string |control-plane/worker       |
=======
|속성           |이름               |타입   |KEY|비고                               |
|---            |---                |---    |---|---                                |
|name           |노드명             |string |NN |mcis vm 이름과 동일                |
|uid            |노드 uid           |string |PK |입력시 값 생성                     |
|role           |역할               |string |NN |control-plane/worker               |
|spec           |spec               |string |NN |                                   |
|public-ip      |공인 IP            |string |   |                                   |
|cluster-uid    |클러스터 uid       |string |FK |cluster foreign-key                |

>>>>>>> Add base source code
=======
|속성           |이름               |타입   |비고                        |
=======
|속성           |이름               |타입   |비고                      |
>>>>>>> Fix json in model.Cluster, model.Node
|---            |---                |---    |---|---                |
|name           |노드명             |string |mcis vm 이름과 동일        |
|credential    |private key       |string |                         |
|publicIp      |공인 IP            |string |                          |
|uid            |노드 uid           |string |uuid                     |
|role           |역할               |string |control-plane/worker    |



>>>>>>> Update entity
