package main

import (
	"fmt"
	"github.com/gshiftlabs-gwedow/godynamo/conf"
	"github.com/gshiftlabs-gwedow/godynamo/conf_file"
	conf_iam "github.com/gshiftlabs-gwedow/godynamo/conf_iam"
	delete_item "github.com/gshiftlabs-gwedow/godynamo/endpoints/delete_item"
	keepalive "github.com/gshiftlabs-gwedow/godynamo/keepalive"
	"github.com/gshiftlabs-gwedow/godynamo/types/attributevalue"
	"log"
	"net/http"
	"os"
)

func main() {
	conf_file.Read()
	conf.Vals.ConfLock.RLock()
	if conf.Vals.Initialized == false {
		panic("the conf.Vals global conf struct has not been initialized")
	}

	// launch a background poller to keep conns to aws alive
	if conf.Vals.Network.DynamoDB.KeepAlive {
		log.Printf("launching background keepalive")
		go keepalive.KeepAlive([]string{})
	}

	// deal with iam, or not
	if conf.Vals.UseIAM {
		iam_ready_chan := make(chan bool)
		go conf_iam.GoIAM(iam_ready_chan)
		_ = <-iam_ready_chan
	}
	conf.Vals.ConfLock.RUnlock()

	tn := "test-godynamo-livetest"
	tablename1 := tn
	fmt.Printf("tablename1: %s\n", tablename1)

	// DELETE AN ITEM
	del_item1 := delete_item.NewDeleteItem()
	del_item1.TableName = tablename1
	del_item1.Key["TheHashKey"] = &attributevalue.AttributeValue{S: "AHashKey1"}
	del_item1.Key["TheRangeKey"] = &attributevalue.AttributeValue{N: "1"}

	body, code, err := del_item1.EndpointReq()
	if err != nil || code != http.StatusOK {
		fmt.Printf("fail delete %d %v %s\n", code, err, string(body))
		os.Exit(1)
	}
	fmt.Printf("%v\n%v\n,%v\n", string(body), code, err)
}
