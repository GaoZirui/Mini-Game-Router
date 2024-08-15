# Diff Details

Date : 2024-08-06 16:18:19

Directory /data/home/ziruigao/src/mini-game-router-v1

Total : 47 files,  498 codes, 0 comments, 80 blanks, all 578 lines

[Summary](results.md) / [Details](details.md) / [Diff Summary](diff.md) / Diff Details

## Files
| filename | language | code | comment | blank | total |
| :--- | :--- | ---: | ---: | ---: | ---: |
| [core/balancer/balancerPrototype.go](/core/balancer/balancerPrototype.go) | Go | 55 | 0 | 8 | 63 |
| [core/balancer/consistentHashBalancer.go](/core/balancer/consistentHashBalancer.go) | Go | 91 | 0 | 19 | 110 |
| [core/balancer/dynamicBalancer.go](/core/balancer/dynamicBalancer.go) | Go | 55 | 0 | 15 | 70 |
| [core/balancer/myPicker.go](/core/balancer/myPicker.go) | Go | 93 | 0 | 12 | 105 |
| [core/balancer/randomBalancer.go](/core/balancer/randomBalancer.go) | Go | 42 | 0 | 14 | 56 |
| [core/balancer/staticBalancer.go](/core/balancer/staticBalancer.go) | Go | 52 | 0 | 15 | 67 |
| [core/balancer/weightBalancer.go](/core/balancer/weightBalancer.go) | Go | 60 | 0 | 17 | 77 |
| [core/config/config.go](/core/config/config.go) | Go | 133 | 0 | 23 | 156 |
| [core/etcd/discovery.go](/core/etcd/discovery.go) | Go | 72 | 0 | 14 | 86 |
| [core/etcd/register.go](/core/etcd/register.go) | Go | 74 | 0 | 10 | 84 |
| [core/init/init.go](/core/init/init.go) | Go | 19 | 0 | 4 | 23 |
| [core/metrics/clientMetrics.go](/core/metrics/clientMetrics.go) | Go | 48 | 0 | 13 | 61 |
| [core/metrics/serverMetrics.go](/core/metrics/serverMetrics.go) | Go | 34 | 0 | 11 | 45 |
| [core/netToolkit/netToolkit.go](/core/netToolkit/netToolkit.go) | Go | 86 | 0 | 18 | 104 |
| [core/resolver/myResolver.go](/core/resolver/myResolver.go) | Go | 70 | 0 | 13 | 83 |
| [core/router/router.go](/core/router/router.go) | Go | 161 | 0 | 23 | 184 |
| [core/tools/randomPickMap.go](/core/tools/randomPickMap.go) | Go | 52 | 0 | 13 | 65 |
| [demo/client/client.go](/demo/client/client.go) | Go | 103 | 0 | 15 | 118 |
| [demo/control/control.go](/demo/control/control.go) | Go | 41 | 0 | 8 | 49 |
| [demo/server/server.go](/demo/server/server.go) | Go | 75 | 0 | 15 | 90 |
| [go.mod](/go.mod) | Go Module File | 35 | 0 | 6 | 41 |
| [go.sum](/go.sum) | Go Checksum File | 103 | 0 | 1 | 104 |
| [proto/hello.pb.go](/proto/hello.pb.go) | Go | 185 | 0 | 26 | 211 |
| [proto/hello.proto](/proto/hello.proto) | Protocol Buffers | 12 | 0 | 4 | 16 |
| [proto/hello_grpc.pb.go](/proto/hello_grpc.pb.go) | Go | 95 | 0 | 16 | 111 |
| [test/test.go](/test/test.go) | Go | 40 | 0 | 3 | 43 |
| [/data/home/ziruigao/src/mini-game-router/cmd/mini-game-router/main.go](//data/home/ziruigao/src/mini-game-router/cmd/mini-game-router/main.go) | Go | -29 | 0 | -9 | -38 |
| [/data/home/ziruigao/src/mini-game-router/core/balancer/balancerPrototype.go](//data/home/ziruigao/src/mini-game-router/core/balancer/balancerPrototype.go) | Go | -39 | 0 | -6 | -45 |
| [/data/home/ziruigao/src/mini-game-router/core/balancer/consistentHashBalancer.go](//data/home/ziruigao/src/mini-game-router/core/balancer/consistentHashBalancer.go) | Go | -83 | 0 | -18 | -101 |
| [/data/home/ziruigao/src/mini-game-router/core/balancer/myBalancer.go](//data/home/ziruigao/src/mini-game-router/core/balancer/myBalancer.go) | Go | -98 | 0 | -14 | -112 |
| [/data/home/ziruigao/src/mini-game-router/core/balancer/randomBalancer.go](//data/home/ziruigao/src/mini-game-router/core/balancer/randomBalancer.go) | Go | -56 | 0 | -17 | -73 |
| [/data/home/ziruigao/src/mini-game-router/core/balancer/weightBalancer.go](//data/home/ziruigao/src/mini-game-router/core/balancer/weightBalancer.go) | Go | -73 | 0 | -20 | -93 |
| [/data/home/ziruigao/src/mini-game-router/core/config/config.go](//data/home/ziruigao/src/mini-game-router/core/config/config.go) | Go | -51 | 0 | -12 | -63 |
| [/data/home/ziruigao/src/mini-game-router/core/etcd/discovery.go](//data/home/ziruigao/src/mini-game-router/core/etcd/discovery.go) | Go | -128 | 0 | -23 | -151 |
| [/data/home/ziruigao/src/mini-game-router/core/etcd/register.go](//data/home/ziruigao/src/mini-game-router/core/etcd/register.go) | Go | -74 | 0 | -11 | -85 |
| [/data/home/ziruigao/src/mini-game-router/core/proxy/redirector/redirector.go](//data/home/ziruigao/src/mini-game-router/core/proxy/redirector/redirector.go) | Go | -46 | 0 | -11 | -57 |
| [/data/home/ziruigao/src/mini-game-router/core/proxy/sidecar/sidecar.go](//data/home/ziruigao/src/mini-game-router/core/proxy/sidecar/sidecar.go) | Go | -65 | 0 | -18 | -83 |
| [/data/home/ziruigao/src/mini-game-router/core/router/router.go](//data/home/ziruigao/src/mini-game-router/core/router/router.go) | Go | -34 | 0 | -6 | -40 |
| [/data/home/ziruigao/src/mini-game-router/core/stateful/stateful.go](//data/home/ziruigao/src/mini-game-router/core/stateful/stateful.go) | Go | -58 | 0 | -11 | -69 |
| [/data/home/ziruigao/src/mini-game-router/demo/client/client.go](//data/home/ziruigao/src/mini-game-router/demo/client/client.go) | Go | -57 | 0 | -10 | -67 |
| [/data/home/ziruigao/src/mini-game-router/demo/server/server.go](//data/home/ziruigao/src/mini-game-router/demo/server/server.go) | Go | -60 | 0 | -14 | -74 |
| [/data/home/ziruigao/src/mini-game-router/go.mod](//data/home/ziruigao/src/mini-game-router/go.mod) | Go Module File | -34 | 0 | -5 | -39 |
| [/data/home/ziruigao/src/mini-game-router/go.sum](//data/home/ziruigao/src/mini-game-router/go.sum) | Go Checksum File | -102 | 0 | -1 | -103 |
| [/data/home/ziruigao/src/mini-game-router/proto/hello.pb.go](//data/home/ziruigao/src/mini-game-router/proto/hello.pb.go) | Go | -185 | 0 | -26 | -211 |
| [/data/home/ziruigao/src/mini-game-router/proto/hello.proto](//data/home/ziruigao/src/mini-game-router/proto/hello.proto) | Protocol Buffers | -12 | 0 | -4 | -16 |
| [/data/home/ziruigao/src/mini-game-router/proto/hello_grpc.pb.go](//data/home/ziruigao/src/mini-game-router/proto/hello_grpc.pb.go) | Go | -95 | 0 | -16 | -111 |
| [/data/home/ziruigao/src/mini-game-router/utils/logger/logger.go](//data/home/ziruigao/src/mini-game-router/utils/logger/logger.go) | Go | -9 | 0 | -4 | -13 |

[Summary](results.md) / [Details](details.md) / [Diff Summary](diff.md) / Diff Details