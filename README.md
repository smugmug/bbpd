## bbpd: A User's Guide

### Introduction

*bbpd* is an http proxy for Amazon's DynamoDB service.

For an overview of what DynamoDB is, please see:

http://aws.amazon.com/dynamodb/

To install *bbpd*, run the following command:

        go get github.com/smugmug/bbpd

*bbpd* is written in Go, and requires a Go 1.1 or higher toolchain to be installed on your system
if you want to build it. If you just want to run it, then use apt-get as described above.

If you want to hack on bbpd, you will need a Go environment.
To understand how to use Go code in your environment, please see:

        http://golang.org/doc/install

and other documentation on the golang.org site.

### Configuration

**Important!**

*bbpd* requires the configuration file used by its underlying library, GoDynamo. Please see
the GoDynamo documentation (available at https://github.com/smugmug/godynamo/blob/master/README.md) for
documentation regarding this configuration file. **You MUST properly configure GoDynamo for *bbpd* to
function correctly.**


### Running

When installed via `go get`, `bbpd` will reside in your `$GOPATH/bin` directory, and you should
make sure that is in your shell's executable search path (`$PATH` etc), or you may copy the
executable to an alternate location.

This package also includes some convenience scripts for managing `bbpd` as a daemon which
you may wish to use or alter.

In the `bin/bbpd` directory are two shell scripts: `bbpd_daemon` and `bbpd_ctl`. Call
`bbpd_ctl` with arguments `start` `stop` or `status`. These scripts assume `bbpd` has been copied into
`/usr/bin`. These are useful if you want to avoid upstart (they are like old apachectl etc).

### Use

The `curl` utility is used for examples below as it tends to be available for most platforms.

First make sure that bbpd is running:

        curl "http://localhost:12333/Status"

You should see some output. To make this more readable, add the `compact` and `indent` options:

        curl "http://localhost:12333/Status?indent=1&compact=1"

Which produces a more readable list of available endpoints. You will note that if you omit `compact`, you
will get some more information including a measure of the duration of the request, which might be useful
for benchmarking.

Here is an example using GetItem

        curl -X POST -d '{"TableName":"mytable","Key":{"Date":{"N":"20131001"},"UserID":{"N":"1"}}}' "http://localhost:12333/GetItem?indent=1&compact=1"

Other endpoints work similarly - you name the endpoint to be called, and provide a JSON serialization of the request you wish
to submit. `bbpd` takes care of adding authorization and other headers for you.

There is also a "compatibility mode" which makes the choice of endpoint a header, as described in the AWS documentation.
In this mode, we could call GetItem like:

        curl -H "X-Amz-Target: DynamoDB_20120810.GetItem" -X POST -d '{"TableName":"mytable","Key":{"Date":{"N":"20131001"},"UserID":{"N":"1"}}}' "http://localhost:12333/"

The "/" route is reserved for these "compatibility mode" endpoint.

Other endpoints are accessed similarly. See the AWS documentation for specific request structure.

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

`AttrbiuteValue`:

        {"AStringSet":{"SS":["a","b","c"]}}
        {"ANumber":{"N":"4"}}
        {"AList":[{"N":"4"},{"SS":["a","b","c"]}]}

are translated to these basic JSON values:

        {"AStringSet":["a","b","c"]}
        {"ANumber":4}
        {"AList":[4,["a","b","c"]]}

`bbpd` now includes support for passing in basic JSON documents in place of Items in the
following endpoints:

- GetItemJSON
- PutItemJSON
- BatchGetItemJSON
- BatchWriteItemJSON

These are not AWS endpoints so they must be called explicitly (there are no `X-Amz-Target`
designations for these, you cannot simply POST your input to the default toplevel route).

They are called as

        http://localhost:$PORT/GetItemJSON
        http://localhost:$PORT/GetItemJSON
        http://localhost:$PORT/BatchGetItemJSON
        http://localhost:$PORT/BatchWriteItemJSON

Note that AWS itself does not support basic JSON - the support is always delivered by a
coercion of basic JSON to and from `AttrbiuteValue`. This coercion is lossy! For example,
a `B` or `BS` will be coerced to a string type (`S`, `SS`) and `NULL` types will be
coerced to `BOOL`. Use with caution.

This feature is only enabled for `Item` types, not for `Key` or other `AttributeValue`
aliases. So for example, `BatchWriteItemJSON` requests of type `DeleteRequest` cannot use
basic JSON, only `PutRequest`.

Here is a quick illustration that shows the same PutItem request using both AttributeValues
and basic JSON:

        curl -H "X-Amz-Target: DynamoDB_20120810.PutItem" -X POST -d '{"TableName":"test-godynamo-livetest","Item":{"TheHashKey":{"S":"a-hash-key"},"TheRangeKey":{"N":"1"},"num":{"N":"1"},"numlist":{"NS":["1","2","3","-7234234234.234234"]},"stringlist":{"SS":["pk1_a","pk1_b","pk1_c"]}}}' http://localhost:12333/;

        curl -X POST -d '{"TableName":"test-godynamo-livetest","Item":{"TheHashKey":"a-hash-key","TheRangeKey":1,"num":1,"numlist":[1,2,3,9,-7234234234.234234],"stringlist":["pk1_a","pk1_b","pk1_c"]}}' http://localhost:12333/PutItemJSON;

See the `tests` directory for more examples.
