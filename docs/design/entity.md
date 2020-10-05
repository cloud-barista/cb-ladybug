# Entity

## Cluster
> 클러스터 정보

|속성           |이름               |타입   |KEY|비고                               |
|---            |---                |---    |---|---                                |
|name           |클러스터 명        |string |NN |                                   |
|uid            |클러스터 uid       |string |PK |입력시 값 생성                     |
|namespace      |MCIS 네임스페이스  |string |NN |                                   |
|mcis           |MCIS 명            |string |NN |                                   |
|cluster-config |클러스터 연결정보  |string |   |Kubernetes 인 경우 kubeconfig.yaml |



## Node
> 클러스터의 노드 정보

|속성           |이름               |타입   |KEY|비고                               |
|---            |---                |---    |---|---                                |
|name           |노드명             |string |NN |mcis vm 이름과 동일                |
|uid            |노드 uid           |string |PK |입력시 값 생성                     |
|role           |역할               |string |NN |control-plane/worker               |
|spec           |spec               |string |NN |                                   |
|public-ip      |공인 IP            |string |   |                                   |
|cluster-uid    |클러스터 uid       |string |FK |cluster foreign-key                |

