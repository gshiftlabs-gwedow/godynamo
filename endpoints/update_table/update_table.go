// Support for the DynamoDB UpdateTable endpoint.
//
// example use:
//
// tests/update_table-livestest.go
//
package update_table

import (
	"encoding/json"
	"github.com/gshiftlabs-gwedow/godynamo/authreq"
	"github.com/gshiftlabs-gwedow/godynamo/aws_const"
	create_table "github.com/gshiftlabs-gwedow/godynamo/endpoints/create_table"
	"github.com/gshiftlabs-gwedow/godynamo/types/globalsecondaryindex"
	"github.com/gshiftlabs-gwedow/godynamo/types/provisionedthroughput"
)

const (
	ENDPOINT_NAME        = "UpdateTable"
	UPDATETABLE_ENDPOINT = aws_const.ENDPOINT_PREFIX + ENDPOINT_NAME
)

type UpdateTable struct {
	GlobalSecondaryIndexUpdates *globalsecondaryindex.GlobalSecondaryIndexUpdates `json:",omitempty"`
	TableName                   string
	ProvisionedThroughput       *provisionedthroughput.ProvisionedThroughput `json:",omitempty"`
}

func NewUpdateTable() *UpdateTable {
	update_table := new(UpdateTable)
	update_table.GlobalSecondaryIndexUpdates =
		globalsecondaryindex.NewGlobalSecondaryIndexUpdates()
	update_table.ProvisionedThroughput =
		provisionedthroughput.NewProvisionedThroughput()
	return update_table
}

// Update is an alias for backwards compatibility
type Update UpdateTable

func NewUpdate() *Update {
	update_table := NewUpdateTable()
	update := Update(*update_table)
	return &update
}

type Request UpdateTable

type Response create_table.Response

func NewResponse() *Response {
	cr := create_table.NewResponse()
	r := Response(*cr)
	return &r
}

func (update_table *UpdateTable) EndpointReq() ([]byte, int, error) {
	// returns resp_body,code,err
	reqJSON, json_err := json.Marshal(update_table)
	if json_err != nil {
		return nil, 0, json_err
	}
	return authreq.RetryReqJSON_V4(reqJSON, UPDATETABLE_ENDPOINT)
}

func (update *Update) EndpointReq() ([]byte, int, error) {
	update_table := UpdateTable(*update)
	return update_table.EndpointReq()
}

func (req *Request) EndpointReq() ([]byte, int, error) {
	update_table := UpdateTable(*req)
	return update_table.EndpointReq()
}
