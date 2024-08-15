# Quick Start
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

$ docker-compose exec control ./control
$ docker-compose exec client ./client --userNum=1000 --requestNum=1000 --sleepTime=50
```