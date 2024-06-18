package grpc

import (
	"fmt"

	"github.com/knstch/course/internal/app/config"
	"github.com/knstch/course/internal/app/grpc/grpcvideo"
	"google.golang.org/grpc"
)

type GrpcClient struct {
	Client grpcvideo.VideoServiceClient
}

func NewGrpcClient(config *config.Config) (*GrpcClient, error) {
	conn, err := grpc.NewClient(fmt.Sprintf("%v:%v", config.CdnGrpcHost, config.CdnGrpcPort), grpc.WithInsecure(), grpc.WithBlock(), grpc.WithDefaultCallOptions(grpc.MaxCallSendMsgSize(1073741824)))
	if err != nil {
		return nil, err
	}

	return &GrpcClient{
		Client: grpcvideo.NewVideoServiceClient(conn),
	}, nil
}
