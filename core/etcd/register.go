package etcd

import (
	"context"
	"fmt"
	"runtime"
	"time"
	"ziruigao/mini-game-router/core/config"
	"ziruigao/mini-game-router/core/metrics"
	"ziruigao/mini-game-router/core/router"

	"github.com/rs/zerolog/log"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/disk"
	"github.com/shirou/gopsutil/mem"
	clientv3 "go.etcd.io/etcd/client/v3"
)

type ServiceRegister struct {
	cli           *clientv3.Client
	leaseID       clientv3.LeaseID
	keepAliveChan <-chan *clientv3.LeaseKeepAliveResponse
	key           string
	val           string
	serverMetrics *metrics.ServerMetrics
}

type ServiceRegisterOpts struct {
	EtcdConfig     *config.EtcdConfig
	EtcdLease      int64
	Endpoint       *router.Endpoint
	ServiceMetrics *metrics.ServerMetrics
}

func NewServiceRegister(opts *ServiceRegisterOpts) (*ServiceRegister, error) {
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   opts.EtcdConfig.Endpoints,
		DialTimeout: opts.EtcdConfig.DialTimeout,
		Username:    opts.EtcdConfig.Username,
		Password:    opts.EtcdConfig.Password,
	})
	if err != nil {
		log.Fatal().Msg(err.Error())
	}
	ser := &ServiceRegister{
		cli:           cli,
		key:           opts.Endpoint.Namespace + "/" + opts.Endpoint.Name + "/" + opts.Endpoint.ToAddr(),
		val:           opts.Endpoint.ToString(),
		serverMetrics: opts.ServiceMetrics,
	}

	//申请租约设置时间keepalive
	if err := ser.putKeyWithLease(opts.EtcdLease); err != nil {
		return nil, err
	}

	return ser, nil
}

func (s *ServiceRegister) putKeyWithLease(lease int64) error {
	go func() {
		ticker := time.NewTicker(time.Second * 5)
		for range ticker.C {
			s.uploadPerformance()
		}
	}()
	//设置租约时间
	resp, err := s.cli.Grant(context.Background(), lease)
	if err != nil {
		return err
	}
	//注册服务并绑定租约
	_, err = s.cli.Put(context.Background(), s.key, s.val, clientv3.WithLease(resp.ID))
	if err != nil {
		return err
	}
	//设置续租 定期发送需求请求
	leaseRespChan, err := s.cli.KeepAlive(context.Background(), resp.ID)
	if err != nil {
		log.Info().Msg(err.Error())
	}
	s.leaseID = resp.ID
	s.keepAliveChan = leaseRespChan

	log.Info().Msg(fmt.Sprintf("put key: %s val: %s success\n", s.key, s.val))
	return nil
}

// ListenLeaseRespChan 监听 续租情况
func (s *ServiceRegister) ListenLeaseRespChan() {
	for leaseKeepResp := range s.keepAliveChan {
		log.Debug().Msg(fmt.Sprintf("set lease success: %v", leaseKeepResp))
	}
	log.Info().Msg("close lease")
}

// Close 注销服务
func (s *ServiceRegister) Close() error {
	//撤销租约
	if _, err := s.cli.Revoke(context.Background(), s.leaseID); err != nil {
		return err
	}
	log.Info().Msg("revoke lease")
	return s.cli.Close()
}

func (s *ServiceRegister) uploadPerformance() {
	leaseResp, err := s.cli.Grant(context.Background(), 10) // 设置过期时间为10秒
	if err != nil {
		log.Fatal().Msg(err.Error())
	}

	key := "performance/" + s.key

	cpuPercent, _ := cpu.Percent(time.Second, false)
	memInfo, _ := mem.VirtualMemory()
	parts, _ := disk.Partitions(true)
	diskInfo, _ := disk.Usage(parts[0].Mountpoint)
	requestInDuration := s.serverMetrics.GetRequestNum()
	s.serverMetrics.ClearRequestNum()

	s.cli.Put(context.Background(), key, (&metrics.ServerPerformance{
		NumGoroutine:      runtime.NumGoroutine(),
		CpuPercent:        cpuPercent[0],
		MemPercent:        memInfo.UsedPercent,
		DiskPercent:       diskInfo.UsedPercent,
		RequestInDuration: requestInDuration,
	}).ToString(), clientv3.WithLease(leaseResp.ID))
}
