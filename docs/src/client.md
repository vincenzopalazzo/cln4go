# cln4go client

At the current state of the library the client provided by cln4go 
is really minimal and it do not provide a typed module with all the types 
of core lightning API, this feature is left for the future when a 
modular code generator will be ready.

For the moment the definition of the types are left to the user of this library,
to give also the flexibility to define only the model that the user needs.
However, with the go generics feature this library is still strongly typed, 
and allow to make typed calls to core lightning.

## Basic Usage of the Client

The client is implemented with a procedural programming style due the limitation of 
go generics that do not allow to have generics paramters on a single method of a struct.

So, a basic usage of the client is

```golang
package main

import (
    "fmt",
    "os",
    
    cln "github.com/vincenzopalazzo/cln4go/client"
)

var rpc *cln.UnixRPC

type Map = map[string]any

func init() {
    path := os.Getenv("CLN_UNIX_SOCKET")
    if path == "" {
        err := fmt.Errorf("Unix path not exported with the CLN_UNIX_SOCKET env variable")
        panic(err)
    }
    
    rpc, _ = cln.NewUnix(path)
}

func main() {
    getinfo, err := cln.Call[Map, Map](rpc, "getinfo", map[string]any{})
    if err != nil {
        fmt.Printf("cln4go core lightning error: %s", err)
    } else { 
        fmt.Printf("cln4go getinfo: node alias %s", getinfo["alias"])
    }
}
```

## Improve Client performance

One of the pain of golang is the reflection used to encode and decode an object from and to JSON, 
and this can be very downgrading while using the library in a restricted env like a raspberry pi 2/3.

In order to work around this problem, the client is able to change the encoder that it is used 
internally to encode and decode the json, and the definition of this encoder is left to the 
user of the library.

A example of custom encoder that use the library [go-json](https://github.com/goccy/go-json) that made the 
conversion from and to json without reflection is reported below

```golang
package json

import (
    "fmt"

    json "github.com/goccy/go-json"
)

type FastJSON struct{}

func (self *FastJSON) EncodeToByte(obj any) ([]byte, error) {
    return json.Marshal(obj)
}

func (self *FastJSON) EncodeToString(obj any) (*string, error) {
    jsonByte, err := json.Marshal(obj)
    if err != nil {
        return nil, err
    }
    jsonStr := string(jsonByte)
    return &jsonStr, nil
}

func (self *FastJSON) DecodeFromString(jsonStr *string, obj any) error {
    return json.Unmarshal([]byte(*jsonStr), &obj)
}

func (self *FastJSON) DecodeFromBytes(jsonByte []byte, obj any) error {
    if len(jsonByte) == 0 {
        return fmt.Errorf("encoding a null byte array")
    }
    return json.Unmarshal(jsonByte, &obj)
}
```

Now we use the encoder as

```golang
package main

import (
    "fmt",
    "os",
   
    json "./json.go"
   
    cln "github.com/vincenzopalazzo/cln4go/client"
)

var rpc *cln.UnixRPC

type Map = map[string]any

func init() {
    path := os.Getenv("CLN_UNIX_SOCKET")
    if path == "" {
        err := fmt.Errorf("Unix path not exported with the CLN_UNIX_SOCKET env variable")
        panic(err)
    }
    
    rpc, _ = cln.NewUnix(path)
    rpc.SetEncoder(&json.FastJSON{})
}

func main() {
    getinfo, err := cln.Call[Map, Map](rpc, "getinfo", map[string]any{})
    if err != nil {
        fmt.Printf("cln4go core lightning error: %s", err)
    } else { 
        fmt.Printf("cln4go getinfo: node alias %s", getinfo["alias"])
    }
}
```

A benchmark with the popular library [glightning]() is available [here](./bench.md)
