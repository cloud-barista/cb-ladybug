# Cloud Controller Manager for Single-Cloud Type MCKS

## Introduction
In single-cloud type MCKS, you can run CCM(cloud controller manager)
to use load balancers and others in your cluster.

To get the general information about cloud controller manager, visit following sites:
[OCI CCM][1], [AWS CCM][2], [OpenStack CCM][3], and [NCP CCM][4].

## Support

| Cloud Provider | Release Version                    | Installation         |
|----------------|------------------------------------|----------------------|
| AWS            | under development(v0.8.0 or later) | Prerequisites+Auto   |
| OpenStack      | under development(v0.8.0 or later) | Auto                 |
| NCP(VPC)       | under development(v0.8.0 or later) | Prerequisites+Manual |

## Installation

The single-cloud type MCKS try to apply RBAC and CCM manifests, automatically.

BUT, each CCM supports many other options for specific Cloud Provider.

And some CCM requires prerequisites to use it, as follows:


### AWS CCM

#### Prerequsites
You should [create IAM policies and roles][5] for control plane and worker nodes.

When you create a cluster, you should pass the names of roles as parameters: [sample][6].

#### Preparing and Running (automatic)

The single-cloud type MCKS will:

1. set a provider-id as InstanceID for each node by `src/scripts/bootstrap.sh`
2. associates the above roles to instances.
3. create tags in related instances, security groups, and subnets.
4. generate a cloud configuration and create a secret from it: the information can be filled from CB-Spider via connection config.
5. apply [RBAC manifest][7] and [daemonset manifest][8].




### OpenStack

#### Preparing and Running (automatic)

The single-cloud type MCKS will:

1. generate a cloud configuration and create a secret from it: the required information can be filled from CB-Spider via connection config.
2. apply [RBAC manifest 1][9], [RBAC manifest 2][10] and [daemonset manifest][11].



### NCP(NaverCloudPlatform): Only VPC

#### Prerequsites

You should create a cluster by CB-Ladybug and create a subent for a load balancer in VPC related to the cluster.

When creating a cluster, CB-Ladybug will create a VPC, a subnet for servers, and others.


#### Preparing and Running (manual)

The single-cloud type MCKS will:

1. set a provider-id as ServerInstanceNo for each node by `src/scripts/bootstrap.sh`

You should:

1. generate a cloud configuration and create a secret from it, [manually](#how-to-prepare-ncpvpc-ccm).
2. apply [RBAC manifest][12] and [daemonset manifest][13].

##### How to prepare NCP(VPC) CCM

You should:

1. create `cloud-config` as follows:
```bash
$ cat << EOF > cloud-config
[Global]
cluster-name=example-cluster
access-key=abcDeFghIjkhmNoPqrstuvWxyZ
secret-key=1234567890abcDeFghIjkhmNoPqrstuvWxyZ
subnet-no=10000
lb-subnet-no=20000
lb-network-type-code=PUBLIC # PUBLIC(default), PRIVATE
throughput-type-code=SMALL # SMALL(default)
EOF
```

2. create a secret from the `cloud-config` as follows:
```bash
$ kubectl delete secret -n kube-system cloud-config
$ kubectl create secret -n kube-system generic cloud-config --from-file=cloud.conf=cloud-config
```

3. apply the RBAC manifest and the daemonset manifest as follows:
```bash
$ kubectl apply -f https://raw.githubusercontent.com/cloud-barista/cb-ladybug/master/src/scripts/addons/ccm/ncpvpc/clusterrole-service-account.yaml
$ kubectl apply -f https://raw.githubusercontent.com/cloud-barista/cb-ladybug/master/src/scripts/addons/ccm/ncpvpc/ncp-cloud-controller-manager-daemonset.yaml
```

[1]: https://github.com/oracle/oci-cloud-controller-manager/blob/master/README.md
[2]: https://github.com/kubernetes/cloud-provider-aws/
[3]: https://github.com/kubernetes/cloud-provider-openstack/blob/master/docs/openstack-cjloud-controller-manager/using-openstack-cloud-controller-manager.md
[4]: https://github.com/cloud-barista/cloud-provider-ncp/
[5]: https://cloud-provider-aws.sigs.k8s.io/prerequisites/
[6]: https://github.com/cloud-barista/cb-ladybug/blob/master/docs/test/cluster-create-aws.sh
[7]: https://github.com/cloud-barista/cb-ladybug/blob/master/src/scripts/addons/ccm/aws/clusterrole-service-account.yaml
[8]: https://github.com/cloud-barista/cb-ladybug/blob/master/src/scripts/addons/ccm/aws/aws-cloud-controller-manager-daemonset.yaml
[9]: https://github.com/cloud-barista/cb-ladybug/blob/master/src/scripts/addons/ccm/openstack/cloud-controller-manager-roles.yaml
[10]: https://github.com/cloud-barista/cb-ladybug/blob/master/src/scripts/addons/ccm/openstack/cloud-controller-manager-role-bindings.yaml
[11]: https://github.com/cloud-barista/cb-ladybug/blob/master/src/scripts/addons/ccm/openstack/openstack-cloud-controller-manager-ds.yaml
[12]: https://github.com/cloud-barista/cb-ladybug/blob/master/src/scripts/addons/ccm/ncpvpc/clusterrole-service-account.yaml
[13]: https://github.com/cloud-barista/cb-ladybug/blob/master/src/scripts/addons/ccm/ncpvpc/ncp-cloud-controller-manager-daemonset.yaml

