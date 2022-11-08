package service

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"

	"github.com/cloud-barista/cb-mcks/src/core/provision"
	"github.com/cloud-barista/cb-mcks/src/core/spider"
	"github.com/cloud-barista/cb-mcks/src/core/tumblebug"
)

// Imported from cloud-provider-aws/pkg/providers/v2/tags.go
const (
	// TagNameKubernetesClusterPrefix is the tag name we use to differentiate multiple
	// logically independent clusters running in the same AZ.
	// tag format: kubernetes.io/cluster/<clusterID> = shared|owned
	// The tag key = TagNameKubernetesClusterPrefix + clusterID
	// The tag value is an ownership value
	TagNameKubernetesClusterPrefix = "kubernetes.io/cluster/"

	// ResourceLifecycleOwned is the value we use when tagging resources to indicate
	// that the resource is considered owned and managed by the cluster,
	// and in particular that the lifecycle is tied to the lifecycle of the cluster.
	ResourceLifecycleOwned = "owned"
)

func awsPrepareCCM(connectionName, clusterName string, vms []tumblebug.VM, provisioner *provision.Provisioner, cpRole, workerRole string) error {
	for _, vm := range vms {
		role := workerRole
		_, exists := provisioner.ControlPlaneMachines[vm.Name]
		if exists {
			role = cpRole
		}

		if err := awsAssociateIamInstanceProfile(connectionName, vm.CspViewVmDetail.IId.SystemId, role); err != nil {
			return errors.New(fmt.Sprintf("Failed to associate IAM instance profile: %v", err))
		}

		if err := awsCreateTags(connectionName, vm.CspViewVmDetail.IId.SystemId, clusterName, ResourceLifecycleOwned); err != nil {
			return errors.New(fmt.Sprintf("Failed to create tags for id(%s): %v", vm.CspViewVmDetail.IId.SystemId, err))
		}

		for _, sgid := range vm.CspViewVmDetail.SecurityGroupIIds {
			if err := awsCreateTags(connectionName, sgid.SystemId, clusterName, ResourceLifecycleOwned); err != nil {
				return errors.New(fmt.Sprintf("Failed to create tags for id(%s): %v", sgid.SystemId, err))
			}
		}

		if err := awsCreateTags(connectionName, vm.CspViewVmDetail.SubnetIID.SystemId, clusterName, ResourceLifecycleOwned); err != nil {
			return errors.New(fmt.Sprintf("Failed to create tags for id(%s): %v", vm.CspViewVmDetail.SubnetIID.SystemId, err))
		}
	}

	return nil
}

func awsCreateTags(connectionName, resourceId, clusterName, lifeCycle string) error {
	tagName := TagNameKubernetesClusterPrefix + clusterName
	resourceTag := spider.KeyValue{tagName, lifeCycle}
	byteResourceTag, err := json.Marshal(resourceTag)
	if err != nil {
		return errors.New(fmt.Sprintf("Failed to make tag info: %v", err))
	}
	iKVList := []spider.KeyValue{}
	iKVList = append(iKVList, spider.KeyValue{spider.AwsKeyCreateTagsResourceId, resourceId})
	iKVList = append(iKVList, spider.KeyValue{spider.AwsKeyCreateTagsTag, string(byteResourceTag)})
	anycall := spider.NewAnyCall(connectionName, spider.AwsFidCreateTags, iKVList)
	if err := anycall.POST(); err != nil {
		return errors.New(fmt.Sprintf("Failed to call AwsFidCreateTags: %v", err))
	}

	oKeyValueMap := make(map[string]string)
	for _, e := range anycall.ReqInfo.OKeyValueList {
		oKeyValueMap[e.Key] = e.Value
	}

	var exists bool
	var result, reason string

	if result, exists = oKeyValueMap["Result"]; exists == false {
		return errors.New(fmt.Sprintf("Result is missing"))
	}

	if result == "false" {
		if reason, exists = oKeyValueMap["Reason"]; exists == false {
			return errors.New(fmt.Sprintf("Reason: Empty"))
		} else {
			return errors.New(fmt.Sprintf("Reason: %s", reason))
		}
	}

	return nil
}

func awsAssociateIamInstanceProfile(connectionName, instanceId, role string) error {
	iKVList := []spider.KeyValue{}

	iKVList = append(iKVList, spider.KeyValue{spider.AwsKeyAssociateIamInstanceProfileInstanceId, instanceId})
	iKVList = append(iKVList, spider.KeyValue{spider.AwsKeyAssociateIamInstanceProfileRole, role})
	anycall := spider.NewAnyCall(connectionName, spider.AwsFidAssociateIamInstanceProfile, iKVList)
	if err := anycall.POST(); err != nil {
		return errors.New(fmt.Sprintf("Failed to call AwsFidAssociateIamInstanceProfile: %v", err))
	}

	oKeyValueMap := make(map[string]string)
	for _, e := range anycall.ReqInfo.OKeyValueList {
		oKeyValueMap[e.Key] = e.Value
	}

	var exists bool
	var result, reason string

	if result, exists = oKeyValueMap["Result"]; exists == false {
		return errors.New(fmt.Sprintf("Result is missing"))
	}

	if result == "false" {
		if reason, exists = oKeyValueMap["Reason"]; exists == false {
			return errors.New(fmt.Sprintf("Reason: Empty"))
		} else {
			return errors.New(fmt.Sprintf("Reason: %s", reason))
		}
	}

	return nil
}

const awsCloudConfigGlobalTemplate string = `[Global]{{range $key, $val := .}}\n{{$key}}={{$val}}{{end}}`

func awsBuildCloudConfig(connectionName string) (string, error) {
	anycall := spider.NewAnyCall(connectionName, spider.AwsFidGetRegionInfo, nil)
	if err := anycall.POST(); err != nil {
		return "", errors.New(fmt.Sprintf("Failed to call AwsFidGetRegionInfo: %v", err))
	}

	if err := decodeAndDecryptKeyValueList(anycall.ReqInfo.OKeyValueList); err != nil {
		return "", errors.New(fmt.Sprintf("Failed to call AwsFidGetREgionInfo: %v", err))
	}

	oKeyValueMap := make(map[string]string)
	for _, e := range anycall.ReqInfo.OKeyValueList {
		oKeyValueMap[e.Key] = e.Value
	}

	var exists bool
	var config = make(map[string]string)
	if config["Zone"], exists = oKeyValueMap["Zone"]; exists == false {
		return "", errors.New(fmt.Sprintf("Zone is missing"))
	}

	var buf bytes.Buffer
	tplCloudConfig := template.Must(template.New("config").Parse(openstackCloudConfigGlobalTemplate))
	if err := tplCloudConfig.Execute(&buf, config); err != nil {
		return "", errors.New(fmt.Sprintf("Failed to execute the cloud config template: %v", err))
	}

	return buf.String(), nil
}

func openstackPrepareCCM(connectionName, clusterName string, vms []tumblebug.VM, provisioner *provision.Provisioner) error {
	return nil
}

const openstackCloudConfigGlobalTemplate string = `[Global]{{range $key, $val := .}}\n{{$key}}={{$val}}{{end}}`
const openstackCloudConfigLoadBalancerTemplate string = `\n\n[LoadBalancer]{{range $key, $val := .}}\n{{$key}}={{$val}}{{end}}`

func openstackBuildCloudConfig(connectionName string, additional []spider.KeyValue) (string, error) {
	anycall := spider.NewAnyCall(connectionName, spider.OpenstackFidGetConnectionInfo, nil)
	if err := anycall.POST(); err != nil {
		return "", errors.New(fmt.Sprintf("Failed to call OpenstackGetConnectionInfo: %v", err))
	}

	if err := decodeAndDecryptKeyValueList(anycall.ReqInfo.OKeyValueList); err != nil {
		return "", errors.New(fmt.Sprintf("Failed to call OpenstackGetConnectionInfo: %v", err))
	}

	oKeyValueMap := make(map[string]string)
	for _, e := range anycall.ReqInfo.OKeyValueList {
		oKeyValueMap[e.Key] = e.Value
	}

	var exists bool
	var configGlobal = make(map[string]string)
	if configGlobal["auth-url"], exists = oKeyValueMap["IdentityEndpoint"]; exists == false {
		return "", errors.New(fmt.Sprintf("IdentityEndpoint is missing"))
	}
	if configGlobal["username"], exists = oKeyValueMap["Username"]; exists == false {
		return "", errors.New(fmt.Sprintf("Username is missing"))
	}
	if configGlobal["password"], exists = oKeyValueMap["Password"]; exists == false {
		return "", errors.New(fmt.Sprintf("Password is missing"))
	}
	if configGlobal["tenant-id"], exists = oKeyValueMap["ProjectID"]; exists == false {
		return "", errors.New(fmt.Sprintf("ProjectID is missing"))
	}
	if configGlobal["domain-name"], exists = oKeyValueMap["DomainName"]; exists == false {
		return "", errors.New(fmt.Sprintf("DomainName is missing"))
	}

	var bufGlobal, bufLoadBalancer bytes.Buffer

	tplGlobal := template.Must(template.New("configGlobal").Parse(openstackCloudConfigGlobalTemplate))
	if err := tplGlobal.Execute(&bufGlobal, configGlobal); err != nil {
		return "", errors.New(fmt.Sprintf("Failed to execute the cloud config template: %v", err))
	}

	if additional != nil {
		var configLoadBalancer = make(map[string]string)
		for _, e := range additional {
			configLoadBalancer[e.Key] = e.Value
		}

		tplLoadBalancer := template.Must(template.New("configLoadBalancer").Parse(openstackCloudConfigLoadBalancerTemplate))
		if err := tplLoadBalancer.Execute(&bufLoadBalancer, configLoadBalancer); err != nil {
			return "", errors.New(fmt.Sprintf("Failed to execute the cloud config template: %v", err))
		}
	}

	return bufGlobal.String() + bufLoadBalancer.String(), nil
}

const spider_key = "cloud-barista-cb-spider-cloud-ba" // 32 bytes

// from cb-spider/cloud-info-manager/credential-info-manager/CredentialInfoManager.go
func decodeAndDecryptKeyValueList(keyValueList []spider.KeyValue) error {
	for i, kv := range keyValueList {
		var err error
		var byteDecode, byteDecrypt []byte

		if byteDecode, err = base64.StdEncoding.DecodeString(kv.Value); err != nil {
			return err
		}

		if byteDecrypt, err = decrypt([]byte(spider_key), byteDecode); err != nil {
			return err
		}

		kv.Value = string(byteDecrypt)
		keyValueList[i] = kv
	}
	return nil
}

// Imported from cb-spider/cloud-info-manager/credential-info-manager/CredentialInfoManager.go
// decryption with spider key
func decrypt(dec_key, contents []byte) ([]byte, error) {
	if len(contents) < aes.BlockSize {
		err := fmt.Errorf("decryption: " + "contents too short")
		return nil, err
	}

	cipherBlock, err := aes.NewCipher(dec_key)
	if err != nil {
		return nil, err
	}

	initVector := contents[:aes.BlockSize]
	contents = contents[aes.BlockSize:]
	cipherTextFB := cipher.NewCFBDecrypter(cipherBlock, initVector)
	cipherTextFB.XORKeyStream(contents, contents)

	return contents, nil
}
