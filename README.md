## GoDynamo: A User's Guide

### Introduction

GoDynamo is a API for the DynamoDB database (http://aws.amazon.com/documentation/dynamodb/) written in Go.

GoDynamo supports all endpoints, uses AWSv4 request signing, and supports IAM authentication as well
as traditional AWS keys.

To install GoDynamo, run the following command:

        go get github.com/gshiftlabs-gwedow/godynamo

which installs a package that requires the rest of the packages in the library.

Also installed as dependencies are

        https://github.com/gshiftlabs-gwedow/goawsroles

which manages support for IAM roles.

GoDynamo is the foundation of *bbpd*, the http proxy daemon for DynamoDB.
You may find that package here:

        https://github.com/gshiftlabs-gwedow/bbpd

To understand how to use Go code in your environment, please see:

        http://golang.org/doc/install

and other documentation on the golang.org site.

### Configuration

GoDynamo is configured with an external file. This allows you to write programs that do not contain
hardcoded authentication values. The `conf_file` package contains an exported method `Read`
that must be called to read these configuration variables into your program state, where they
will be visible as the exported global variable `Conf`. The `Read` method will first look for
`~/.aws-config.json`, and then `/etc/aws-config.json`. You may also set an environment variable
`GODYNAMO_CONF_FILE` that will permit you to specify a fully-qualified file path for your own
conf file. If none of those files are present,
the `Read` method will return false and it is advised that your program terminate.

A sample of a skeleton `aws-config.json` file is found in `conf/SAMPLE-aws-config.json`.
Please see the go docs for the `conf` package to see an explanation of the fields and their use.
It is recommended that you set file permissions on the configuration file to be as restrictive
as possible.

For convenience, here is the sample configuration file (comments nonstandard):

    {
        "extends":[],
        "services": {
            "default_settings":{
                "params":{
                    // Traditional AWS access/secret authentication pair.
                    "access_key_id":"xxx",
                    "secret_access_key":"xxx",
                    // If you use syslogd (a linux or *bsd system), you may set this to "true".
                    // (currently unusued)
                    "use_sys_log":true
                }
            },
            "dynamo_db": {
                "host":"dynamodb.us-east-1.amazonaws.com",
                "zone":"us-east-1",
                // You can alternately set the scheme/port to be https/443.
                "scheme":"http",
                "port":80,
                // If set to true, programs that are written with godynamo may
                // opt to launch the keepalive goroutine to keep conns open.
                "keepalive":true,
                "iam": {
                    // If you do not want to use IAM (i.e. just use access_key/secret),
                    // set this to false and use the settings above.
                    "use_iam":true,
                    // The role provider is described in the goawsroles package.
                    // See: https://github.com/gshiftlabs-gwedow/goawsroles/
                    // Currently the only support is for the "file" provider, whereby
                    // roles data is written to local files.
                    "role_provider":"file",
                    // The identifier (filename, etc) for the IAM Access Key
                    "access_key":"role_access_key",
                    // The identifier (filename, etc) for the IAM Secret Key
                    "secret_key":"role_secret_key",
                    // The identifier (filename, etc) for the IAM Token
                    "token":"role_token",
                    // If using the "file" role provider, the base dir to read IAM files.
                    "base_dir":"/dir/where/you/update/role_files",
                    // Set to true if you would like the roles resource watched for changes
                    // and automatically (and atomically) updated.
                    "watch":true
                }
            }
        }
    }

In this configuration example, the recommended option of using IAM credentials has been selected,
with the source for these credentials being local text files. Creating the automation to
retrieve these credential files and store them on your host is specific to your installation
as IAM is capable of setting fine-grained permissions. See your sysadmin for assistance. If
you do not wish to use IAM or cannot create the automation to keep your local credential files
up to date, you may wish to set `use_iam` to false and just set the access and secret keypair.


### Some necessary boilerplate

In any program you write using GoDynamo, you must first make sure that your configuration has
been initialized properly. You will optionally wish to use IAM support for authentication.
Below is some boilerplate to enable both of these in your program:

        import (
                "fmt"
                "log"
                conf_iam "github.com/gshiftlabs-gwedow/godynamo/conf_iam"
                keepalive "github.com/gshiftlabs-gwedow/godynamo/keepalive"
                "github.com/gshiftlabs-gwedow/godynamo/conf"
                "github.com/gshiftlabs-gwedow/godynamo/conf_file"
        )

        func main() {

            // Read in the conf file, panic if it hasn't been initialized correctly.
            conf_file.Read()
            if conf.Vals.Initialized == false {
                panic("the conf.Vals global conf struct has not been initialized")
            }

     	    // launch a background poller to keep conns to aws alive
	        if conf.Vals.Network.DynamoDB.KeepAlive {
		        log.Printf("launching background keepalive")
		        go keepalive.KeepAlive([]string{conf.Vals.Network.DynamoDB.URL})
	        }

            // Initialize a goroutine which will watch for changes in the local files
            // we have chosen (in our conf file) to contain our IAM authentication values.
            // It is assumed that another process refreshes these files.
            // If you opt to use plain old AWS authentication pairs, you don't need this.
            if conf.Vals.UseIAM {
                 iam_ready_chan := make(chan bool)
                 go conf_iam.GoIAM(iam_ready_chan)
                 iam_ready := <- iam_ready_chan
                 if iam_ready {
                     fmt.Printf("using iam\n")
                } else {
                     fmt.Printf("not using iam\n")
                }
            }
            // ... rest of code
        }

For more examples that demonstrate how you might wish to use various endpoint libraries, please refer to the
`tests` directory which contains a series of files that are intended to run against AWS, so executing them
will require valid AWS credentials.

### Special Features

One noteworthy feature of GoDynamo is some convenience functions to get around some static limitations of the
default DynamoDB service. In particular, in `endpoints/batch_get_item` you will find a function `DoBatchGet`
which allows an input structure with an arbitrary number of get requests, which are dispatched in segments
and re-assembled. Likewise, in `endpoints/batch_write_item` you will find a function `DoBatchWrite` which
allows an input structure with an arbitrary number of write requests. These functions are provided
as a convenience and do not alter your provisioning model, so be careful.

*Throttling* occurs in DynamoDB operations when AWS wishes to shape traffic to their service.
GoDynamo utilizes the standard *exponential decay* resubmission algorithm as described in
the AWS documentation. While you will see messages regarding the throttling, GoDynamo continues to
retry your request as per the resubmission algorithm.

### JSON Documents

Amazon has been augmenting their SDKs with wrappers that allow the caller to coerce
their Items (both when writing and reading) to what I will refer to as "basic JSON".

Basic JSON is stripped of the type signifiers ('S','NS', etc) that AWS specifies in their
`AttributeValue` specification (http://docs.aws.amazon.com/amazondynamodb/latest/APIReference/API_AttributeValue.html).

For example, the `AttributeValue`

        {"AString":{"S":"this is a string"}}

is translated to this basic JSON:

        {"AString":"this is a string"}

Here are some other examples:

`AttributeValue`:

        {"AStringSet":{"SS":["a","b","c"]}}
        {"ANumber":{"N":"4"}}
        {"AList":[{"N":"4"},{"SS":["a","b","c"]}]}

are translated to these basic JSON values:

        {"AStringSet":["a","b","c"]}
        {"ANumber":4}
        {"AList":[4,["a","b","c"]]}

GoDynamo now includes support for passing in basic JSON documents in place of Items in the
following endpoints:

- GetItem
- PutItem
- BatchGetItem
- BatchWriteItem

In the case of `GetItem` and `BatchGetItem`, the method is to call the methods
`ToResponseItemJSON` and `ToResponseItemsJSON` respective methods on the `Response`
types for each package, unmarshaled from the response strings returned from AWS
after calling `EndpointReq` to complete requests.

In the case of `PutItem`, use `NewPutItemJSON()` to initialize a variable of type
`PutItemJSON`. The `Item` field of this struct is an `interface{}`, to which you
can assign data representing basic JSON. Once done, call the `ToPutItem()` method on the `PutItemJSON`
variable to translate the basic JSON into `AttributeValue` types and return a `PutItem`
type that you can now call `EndpointReq` on.

In the case of `PutItem`, use `NewBatchWriteItemJSON()` to initialize a variable of type
`BatchWriteItemJSON`. The `Item` field of this `PutRequest` subfield is an `interface {}` you 
can assign data representing basic JSON. Once done, call the `ToBatchWriteItem()` method on the `BatchWriteItemJSON`
variable to translate the basic JSON into `AttributeValue` types and return a `BatchWriteItem`
type that you can now call `EndpointReq` on.

Note that AWS itself does not support basic JSON - the support is always delivered by a
coercion of basic JSON to and from `AttributeValue`. This coercion is lossy! For example,
a `B` or `BS` will be coerced to a string type (`S`, `SS`) and `NULL` types will be
coerced to `BOOL`. Use with caution.

This feature is only enabled for `Item` types, not for `Key` or other `AttributeValue`
aliases. So for example, `BatchWriteItemJSON` requests of type `DeleteRequest` cannot use
basic JSON, only `PutRequest`.

### Troubleshooting

GoDynamo provides verbose error messages when appropriate, as well as STDERR messaging. If error
reporting is not useful, it is possible that DynamoDB itself has a new or changed feature that is
not reflected in GoDynamo. 

### Contact Us

Please contact opensource@gshiftlabs-gwedow.com for information related to this package. 
Pull requests also welcome!

