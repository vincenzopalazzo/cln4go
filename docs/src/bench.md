# cln4go Benchmarks

There is already a popular client in Go that is [glightning](https://github.com/niftynei/glightning), 
so why I need to use another library, where i need to specify the model by hand?

What is the benefit of this?

## cln4go Benefits

The benefit of cln4go is the modularity and the flexibility that it give to the user
to configure the Go client that the user want for core lightning.

One of the flexibility that the API give to use it to inject a custom tracer (logger) 
and a custom JSON encoder. This feature will give to inject a custom JSON encoder
and to use a custom logger that do not interact with the core lightning logger, but
use a different stream such as a File.

See how to implement a custom Encoder in [plugin docs](./plugin.md) and how to write a 
custom logger in the [common docs](./common.md)

In addition, this client do not exclude the possibility to have a library that include
all the core lightning model and call by generating the typed client on top of the cln4go client.

It is on the road map of the library to have a autogenerate client from the core lightning json 
schema, so if you are planning to help with this client, you can consider to use the core lightning
core generation code to generate the strongly typed library on top of the cln4go client.

## Benchmarks

In this section we run benchmarks to compare the cln4go with the 
encoder discussed in the section regarding the [client](./client.md) compared 
with the solution provided by [glightning](https://github.com/niftynei/glightning) library.

TODO put here the figure!

To run yourself the benchmarks, you can run the following commands

```
>> export CLN_UNIX_SOCKET=/run/media/vincent/VincentSSD/.lightning/testnet/lightning-rpc
>> make dep
>> make bench_check

cd bench; go run main.go
2023/04/13 13:51:21 INFO ----------------- cln4go-listnodes -----------------
2023/04/13 13:51:21 INFO Number of run: 1000000000
2023/04/13 13:51:21 INFO Time taken: 28.793444ms
2023/04/13 13:51:21 INFO Size result: 1
2023/04/13 13:51:22 INFO ----------------- glightning-listnodes -----------------
2023/04/13 13:51:22 INFO Number of run: 1000000000
2023/04/13 13:51:22 INFO Time taken: 68.307247ms
2023/04/13 13:51:22 INFO Size result: 2
2023/04/13 13:51:23 INFO ----------------- cln4go-listchannels -----------------
2023/04/13 13:51:23 INFO Number of run: 1000000000
2023/04/13 13:51:23 INFO Time taken: 104.217164ms
2023/04/13 13:51:23 INFO Size result: 3
2023/04/13 13:51:25 INFO ----------------- glightning-listchannels -----------------
2023/04/13 13:51:25 INFO Number of run: 1000000000
2023/04/13 13:51:25 INFO Time taken: 177.613183ms
2023/04/13 13:51:25 INFO Size result: 4
2023/04/13 13:51:25 INFO json: {"bench":[{"name":"cln4go-listnodes","runs":1000000000,"times":28793444,"time_str":"28.793444ms"},{"name":"glightning-listnodes","runs":1000000000,"times":68307247,"time_str":"68.307247ms"},{"name":"cln4go-listchannels","runs":1000000000,"times":104217164,"time_str":"104.217164ms"},{"name":"glightning-listchannels","runs":1000000000,"times":177613183,"time_str":"177.613183ms"}]}
```
