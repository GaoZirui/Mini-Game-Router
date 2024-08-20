cd redis-cluster
bash ./deploy.sh clean
cd ..
sudo rm -rf deploy/grafana-data
sudo rm -rf deploy/.dbdata
sudo rm -rf cmd/demo/client/cpu.pprof
sudo rm -rf cmd/demo-tcp/client/cpu.pprof
sudo rm -rf .VSCodeCounter