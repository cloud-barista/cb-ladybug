# Test 
Test shell scripts 사용법

## Prerequisites 
> 클러스터 생성/삭제 기능을 이용하기 위해서는 Cloud Connection 정보를 등록해야 합니다.

### jq 설치
* shell 에서 json parsing 시 `jq` 유틸리티를 활용합니다.
* https://stedolan.github.io/jq/

```
▒ brew install jq
```

### CB-Spider, CB-Tumblebug 실행

```
▒  docker run -d -p 1024:1024 --name cb-spider cloudbaristaorg/cb-spider:v0.2.0-20200715
▒  docker run -d -p 1323:1323 --name cb-tumblebug --link cb-spider:cb-spider cloudbaristaorg/cb-tumblebug:v0.2.0-20200715
```

### Cloud Connection Info. 등록

####  GCP

* 환경변수 : 클라우드별 연결정보

```
▒ export PROJECT="<project name>"
▒ export PKEY="private key>"
▒ export SA="<service account email>"
```

* 환경변수 : REGION, ZONE

```
▒ export REGION="<region name>" 
▒ export ZONE="<zone name>"

# 예 : asia-northeast3 (서울리전)
▒ export REGION="asia-northeast3" 
▒ export ZONE="asia-northeast3-a"
```

* Cloud Connection Info. 등록

```
▒ ./init.sh GCP
```

* 결과 확인

```
▒ ./get.sh GCP ns,config
```

#### AWS

* 환경변수 : 클라우드별 연결정보

```
▒ export KEY="<aws_access_key_id>"
▒ export SECRET="<aws_secret_access_key>"
```

* 환경변수 : REGION, ZONE

```
▒ export REGION="<region name>" 
▒ export ZONE="<zone name>"

# 예: ap-northeast-1 (일본리전)
▒ export REGION="ap-northeast-1"
▒ export ZONE="ap-northeast-1a"
```

* Cloud Connection Info. 등록

```
▒ ./init.sh AWS
```

* 결과 확인

```
▒ ./get.sh AWS ns,config
```

## Test 

### cb-ladybug 실행

```
▒ export CBLOG_ROOT="$(pwd)"
▒ export CBSTORE_ROOT="$(pwd)"
▒ go run src/main.go
```

### 클러스터 생성
```
▒ /ladybug.sh create [GCP/AWS] <cluster name> <spec:machine-type> <worker-node-count>
```

* 예
```
▒ ./ladybug.sh create GCP cb-cluster n1-standard-2 1   # GCP
▒ ./ladybug.sh create AWS cb-cluster t2.medium 1       # AWS
```

### 클러스터 삭제
```
▒ /ladybug.sh destroy [GCP/AWS] <cluster name>
```

* 예
```
▒ ./ladybug.sh destroy GCP cb-cluster   # GCP
▒ ./ladybug.sh destroy AWS cb-cluster   # AWS
```


## 기타

### SSH key 파일 저장

```
▒ ./savekey.sh [AWS/GCP] <cluster name>
```

* 예
```
▒ ./savekey.sh AWS cb-cluster
▒ cat *.pem
```

### 파일에서 클라우드별 연결정보 얻기

* GCP ( [jq](https://stedolan.github.io/jq/) 설치 필요)

```
▒ source ./env.sh GCP "<json file path>"

# 예
▒ source ./env.sh GCP "${HOME}/.ssh/google-credential-cloudbarista.json"
```

* AWS

```
▒ source ./env.sh AWS "<credentials file path>"

# 예
▒ source ./env.sh AWS "${HOME}/.aws/credentials"
```