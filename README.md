## bbpd: A User's Guide

### Introduction

*bbpd* is an http proxy for Amazon's DynamoDB service.

For an overview of what DynamoDB is, please see:

http://aws.amazon.com/dynamodb/

To install *bbpd*, run the following command:

        go get github.com/smugmug/bbpd

*bbpd* is written in Go, and requires a Go 1.1 or higher toolchain to be installed on your system.

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

In the `bin/bbpd` directory you will find `bbpd.conf` which can be used to start and stop `bbpd`
on Ubuntu systems via `upstart`. Please see Ubuntu documentation for details.

Also in the `bin/bbpd` directory are two shell scripts: `bbpd_daemon` and `bbpd_ctl`. Call
`bbpd_ctl` with arguments `start` `stop` or `status`. These scripts assume `bbpd` has been copied into
`/usr/bin`.

`bbpd` is configured to use ports 12333 and 12334 in that order.

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

### Caveats

This proxy uses the standard `net/http` server. As of Go v1.2 (rc), this server does not support *graceful*
restarts. We are aware of alternative solutions but haven't introduced them to this codebase, hoping
that this feature will shortly be introduced into the standard library. The proxy does have a signal
handler to accept various termination signals, upon which the proxy will stop accepting new
connections and go into a short sleep to allow existing connections to terminate. This isn't an optimal
solution but should be adequate until the standard library catches up.

### Debugging

If you have opted to use `syslog` in your configuration file, you may look in the configured log file for error and
warning messages. Otherwise these are available via `STDERR`.

### Contact Us

Please contact opensource@smugmug.com for information related to this package. 
Pull requests also welcome!
