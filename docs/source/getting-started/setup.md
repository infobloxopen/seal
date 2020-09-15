# Requirements

The following are required for SEAL development:

* docker
* golang 1.12+


# Dev environment

## Clone the repo
Checkout the source code under your GOPATH. This path will depend on how you've installed go. See
the go installation instructions for details on GOPATH. https://golang.org/doc/gopath_code.html

```bash

mkdir -p $GOPATH/src/github.com/infobloxopen
cd $GOPATH/src/github.com/infobloxopen
git clone git@github.com:infobloxopen/seal.git
cd seal
```

## Build, run, and test petstore example
```bash
make petstore
```
