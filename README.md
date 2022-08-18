<div align="center">
  <h1>cln4go</h1>

  <img src="https://preview.redd.it/tcmyd3n69ng41.jpg?width=1999&format=pjpg&auto=webp&s=b79cf22d3e2adcaf52a2d22bcb0568e42eff8bc2" />

  <p>
    <strong> Go library for Core Lightning Daemon with flexible interface </strong>
  </p>

  <span>
   <img alt="GitHub Workflow Status" src="https://img.shields.io/github/workflow/status/vincenzopalazzo/cln4go/Build%20and%20test%20Go?style=flat-square"/>
   <img alt="GitHub go.mod Go version (subdirectory of monorepo)" src="https://img.shields.io/github/go-mod/go-version/vincenzopalazzo/cln4go?filename=plugin%2Fgo.mod&style=flat-square"/>
  </span>

  <h4>
    <a href="https://github.com/vincenzopalazzo/cln4go">Project Homepage</a>
  </h4>
</div>


This repository contains a sequence of libraries that are useful to work with Core Lightning Daemon and develop with Core Lightning using Go.

## Packages

These are the complete list of craters supported right now

| Crate     | Description |
|:----------|:-----------:|
| clng4go-client |    Package that provides means to make RPC bindings from Go code to the core lightning daemon     | 
| cln4go-plugin |    Package that provides a plugin API to give the possibility to implement a plugin in Go     | 
| cln4go-common |    Package that provides common interface for the monorepo. Go     | 
## How to Use
### Core Go Client

```
	path := os.Getenv("CLN_UNIX_SOCKET")
	if path == "" {
		err := fmt.Errorf("Unix path not exported with the CLN_UNIX_SOCKET env variable")
		panic(err)
	}
	client, err := NewUnix(path)
	if err != nil {
		panic(err)
	}
	response, err := Call[UnixRPC, MapReq, GetInfo](client, "getinfo", make(map[string]any, 0))
```
Please look inside the client/unix_test.go for more usage examples. 


### Core Go Plugin 

Please look inside the plugin/example/simple_plugin.go for examples. 

## Contributing guidelines

Read our [Hacking guide](/docs/MAINTAINERS.md)

## Supports

If you want support this library consider to donate with the following methods

- Lightning address: vincenzopalazzo@lntxbot.com
- [Github donation](https://github.com/sponsors/vincenzopalazzo)