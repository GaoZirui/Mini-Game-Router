# Mini-Game-Router
* ziruigao
## Quick Start
```bash
# 构建镜像
$ docker build --rm=true -t mini-game-router-server:latest -f Dockerfile.server .
$ docker build --rm=true -t mini-game-router-client:latest -f Dockerfile.client .
$ docker build --rm=true -t mini-game-router-control:latest -f Dockerfile.control .

# 启动 docker
$ cd deploy
$ docker-compose up

# 初始化设置
$ cd deploy
$ docker-compose exec control ./control

# 启动 3 个服务器
$ cd deploy
$ docker-compose exec server-1 ./server --svrID=server-1 --endpointsNum=5 --showReceive
$ docker-compose exec server-2 ./server --svrID=server-2 --endpointsNum=5 --showReceive
$ docker-compose exec server-3 ./server --svrID=server-3 --endpointsNum=5 --showReceive

# 启动客户端
$ cd deploy
$ docker-compose exec client ./client --userNum=3 --requestNum=2000 --coreNum=4 --debug --countReply

# 控制中心命令
$ cd deploy
# 服务端配置热更新（修改 deploy/serverConfig）
$ docker-compose exec control ./control --op=set-server
# 客户端配置热更新（修改 deploy/clientConfig）
$ docker-compose exec control ./control --op=set-client
# 服务端下线
$ docker-compose exec control ./control --op=close-server
# 服务端上线
$ docker-compose exec control ./control --op=up-server
```
## 非 `docker` 部署指令
* 首先在本地开启 `etcd` 与 `redis`
```bash
# 启动服务端
$ cd deploy
$ go run server.go --svrID=server-1 --configPath=../../../config/serverConfig.yaml --endpointsNum=5 --showReceive
$ go run server.go --svrID=server-2 --configPath=../../../config/serverConfig.yaml --endpointsNum=5 --showReceive
$ go run server.go --svrID=server-3 --configPath=../../../config/serverConfig.yaml --endpointsNum=5 --showReceive

# 初始化配置
$ cd deploy
$ go run control.go --clientConfigPath=../../config/clientConfig.yaml --serverConfigPath=../../config/serverConfig.yaml

# 启动客户端
$ cd deploy
$ go run client.go --configPath=../../../config/clientConfig.yaml --userNum=3 --requestNum=20000 --coreNum=4 --debug --countReply

# 控制中心命令
$ cd deploy
# 服务端配置热更新（修改 config/serverConfig）
$ go run control.go --clientConfigPath=../../config/clientConfig.yaml --serverConfigPath=../../config/serverConfig.yaml --op=set-server
# 客户端配置热更新（修改 config/clientConfig）
$ go run control.go --clientConfigPath=../../config/clientConfig.yaml --serverConfigPath=../../config/serverConfig.yaml --op=set-client
# 服务端下线
$ go run control.go --clientConfigPath=../../config/clientConfig.yaml --serverConfigPath=../../config/serverConfig.yaml --op=close-server
# 服务端上线
$ go run control.go --clientConfigPath=../../config/clientConfig.yaml --serverConfigPath=../../config/serverConfig.yaml --op=up-server
```
## Clean
```bash
bash clean.sh
```
## Clean ETCD
```bash
$ rev=$(etcdctl --endpoints=http://127.0.0.1:2379 endpoint status --write-out="json" | egrep -o '"revision":[0-9]*' | egrep -o '[0-9].*')
$ etcdctl --endpoints=http://127.0.0.1:2379 compact $rev
$ etcdctl --endpoints=http://127.0.0.1:2379 defrag
$ etcdctl endpoint status
$ etcdctl --endpoints=http://127.0.0.1:2379 alarm disarm
```