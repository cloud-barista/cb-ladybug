package spider

import (
	"net/http"
)

func NewAnyCall(connectionName string, fId FUNCTION, iKeyValueList []KeyValue) *AnyCall {
	return &AnyCall{
		ConnectionName: connectionName,
		ReqInfo: AnyCallInfo{
			FID:           fId,
			IKeyValueList: iKeyValueList,
			OKeyValueList: []KeyValue{},
		},
	}
}

func (self *AnyCall) POST() error {
	_, err := self.execute(http.MethodPost, "/anycall", self, &self.ReqInfo)
	if err != nil {
		return err
	}

	return nil
}
