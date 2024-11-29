package main

import (
	"context"
	"flag"
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	extProcPb "github.com/envoyproxy/go-control-plane/envoy/service/ext_proc/v3"
	extproc "gitlab.alibaba-inc.com/cos/extproc/server"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/status"
	klog "k8s.io/klog/v2"
)

var (
	port = flag.Int("port", 9002, "gRPC port")
)

type healthServer struct{}

func (s *healthServer) Check(ctx context.Context, in *grpc_health_v1.HealthCheckRequest) (*grpc_health_v1.HealthCheckResponse, error) {
	klog.Infof("Handling grpc Check request: + %s", in.String())
	return &grpc_health_v1.HealthCheckResponse{Status: grpc_health_v1.HealthCheckResponse_SERVING}, nil
}

func (s *healthServer) Watch(in *grpc_health_v1.HealthCheckRequest, srv grpc_health_v1.Health_WatchServer) error {
	return status.Error(codes.Unimplemented, "Watch is not implemented")
}

func main() {
	flag.Parse()

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		klog.Fatalf("failed to listen: %v", err)
	}

	sopts := []grpc.ServerOption{grpc.MaxConcurrentStreams(1000)}
	s := grpc.NewServer(sopts...)

	extProcPb.RegisterExternalProcessorServer(s, extproc.NewServer())

	grpc_health_v1.RegisterHealthServer(s, &healthServer{})

	klog.Infof("Starting gRPC server on port %s", fmt.Sprintf(":%d", *port))

	var gracefulStop = make(chan os.Signal, 1)
	signal.Notify(gracefulStop, syscall.SIGTERM, syscall.SIGINT)
	go func() {
		sig := <-gracefulStop
		klog.Infof("caught sig: %+v", sig)
		time.Sleep(time.Second)
		klog.Infof("Graceful stop completed")
		os.Exit(0)
	}()
	err = s.Serve(lis)
	if err != nil {
		klog.Fatalf("killing server with %v", err)
	}
}
