package spider

const (
	AwsFidCreateTags           FUNCTION = "createTags"
	AwsKeyCreateTagsResourceId          = "ResourceId"
	AwsKeyCreateTagsTag                 = "Tag"

	AwsFidAssociateIamInstanceProfile           FUNCTION = "associateIamInstanceProfile"
	AwsKeyAssociateIamInstanceProfileInstanceId          = "InstanceId"
	AwsKeyAssociateIamInstanceProfileRole                = "Role"

	AwsFidGetRegionInfo FUNCTION = "getRegionInfo"

	OpenstackFidGetConnectionInfo FUNCTION = "getConnectionInfo"
)

type FUNCTION string

type KeyValue struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type AnyCallInfo struct {
	FID           FUNCTION   `json:"fId"`           // Function ID
	IKeyValueList []KeyValue `json:"iKeyValueList"` // Input Arguments List
	OKeyValueList []KeyValue `json:"oKeyValueList"` // Output Results List
}

type Model struct {
}

type AnyCall struct {
	Model
	ConnectionName string      `json:"connectionName"`
	ReqInfo        AnyCallInfo `json:"reqInfo"`
}
