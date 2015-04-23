package main

import (
	"fmt"
	"github.com/gshiftlabs-gwedow/godynamo/conf"
	"github.com/gshiftlabs-gwedow/godynamo/conf_file"
	conf_iam "github.com/gshiftlabs-gwedow/godynamo/conf_iam"
	list "github.com/gshiftlabs-gwedow/godynamo/endpoints/list_tables"
	keepalive "github.com/gshiftlabs-gwedow/godynamo/keepalive"
	"log"
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

	// List TABLES
	var l list.ListTables
	l.ExclusiveStartTableName = ""
	l.Limit = 100
	lbody, lcode, lerr := l.EndpointReq()
	fmt.Printf("%v\n%v\n,%v\n", string(lbody), lcode, lerr)
}
