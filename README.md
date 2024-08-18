# Mini-Game-Router
* ziruigao
## Quick Start
```bash
$ docker build --rm=true -t mini-game-router-server:latest -f Dockerfile.server .
$ docker build --rm=true -t mini-game-router-client:latest -f Dockerfile.client .
$ docker build --rm=true -t mini-game-router-control:latest -f Dockerfile.control .
$ cd deploy
$ docker-compose up -d

$ docker exec -it redis-1 bash
$ redis-cli --cluster create 127.0.0.1:9001 \
127.0.0.1:9002 \
127.0.0.1:9003 \
127.0.0.1:9004 \
127.0.0.1:9005 \
127.0.0.1:9006 \
--cluster-replicas 1
$ yes

$ cd deploy
$ docker-compose exec control ./control
$ docker-compose exec client ./client --userNum=5000 --requestNum=2000 --coreNum=4
```

## Clean ETCD
```bash
$ rev=$(etcdctl --endpoints=http://127.0.0.1:2379 endpoint status --write-out="json" | egrep -o '"revision":[0-9]*' | egrep -o '[0-9].*')
$ etcdctl --endpoints=http://127.0.0.1:2379 compact $rev
$ etcdctl --endpoints=http://127.0.0.1:2379 defrag
$ etcdctl endpoint status
$ etcdctl --endpoints=http://127.0.0.1:2379 alarm disarm
```

## CMD
```bash
# start server
go run server.go --svrID=server-1 --configPath=../../../config/serverConfig.yaml --endpointsNum=5
go run server.go --svrID=server-2 --configPath=../../../config/serverConfig.yaml --endpointsNum=5
go run server.go --svrID=server-3 --configPath=../../../config/serverConfig.yaml --endpointsNum=5

# init config
go run control.go

# start client
go run client.go --configPath=../../../config/clientConfig.yaml --userNum=3 --requestNum=20000 --coreNum=4 --debug

# update config
go run control.go --op=set-server
go run control.go --op=set-client

# close server
go run control.go --op=close-server

# up server
go run control.go --op=up-server
```