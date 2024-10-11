package client

import (
	"context"
	intrv1 "geektime/webook/api/proto/gen/intr/v1"
	"github.com/ecodeclub/ekit/syncx/atomicx"
	"google.golang.org/grpc"
	"math/rand"
)

// GrayScaleInteractiveClient
// 集成grpc和本地的客户端，进行流量控制
type GrayScaleInteractiveClient struct {
	remote intrv1.InteractiveServiceClient
	local  intrv1.InteractiveServiceClient

	//阈值，选择grpc还是本地调用
	threshold *atomicx.Value[int32]
}

func NewGrayScaleInteractiveClient(remote intrv1.InteractiveServiceClient, local intrv1.InteractiveServiceClient) *GrayScaleInteractiveClient {
	return &GrayScaleInteractiveClient{
		remote:    remote,
		threshold: atomicx.NewValue[int32](),
		local:     local,
	}
}

func (i *GrayScaleInteractiveClient) IncrReadCnt(ctx context.Context, in *intrv1.IncrReadCntRequest, opts ...grpc.CallOption) (*intrv1.IncrReadCntResponse, error) {
	return i.selectClient().IncrReadCnt(ctx, in, opts...)
}

func (i *GrayScaleInteractiveClient) Like(ctx context.Context, in *intrv1.LikeRequest, opts ...grpc.CallOption) (*intrv1.LikeResponse, error) {
	return i.selectClient().Like(ctx, in, opts...)
}

func (i *GrayScaleInteractiveClient) CancelLike(ctx context.Context, in *intrv1.CancelLikeRequest, opts ...grpc.CallOption) (*intrv1.CancelLikeResponse, error) {
	return i.selectClient().CancelLike(ctx, in, opts...)
}

func (i *GrayScaleInteractiveClient) Collect(ctx context.Context, in *intrv1.CollectRequest, opts ...grpc.CallOption) (*intrv1.CollectResponse, error) {
	return i.selectClient().Collect(ctx, in, opts...)
}

func (i *GrayScaleInteractiveClient) Get(ctx context.Context, in *intrv1.GetRequest, opts ...grpc.CallOption) (*intrv1.GetResponse, error) {
	return i.selectClient().Get(ctx, in, opts...)
}

func (i *GrayScaleInteractiveClient) GetByIds(ctx context.Context, in *intrv1.GetByIdsRequest, opts ...grpc.CallOption) (*intrv1.GetByIdsResponse, error) {
	return i.selectClient().GetByIds(ctx, in, opts...)
}

func (i *GrayScaleInteractiveClient) selectClient() intrv1.InteractiveServiceClient {
	// [0, 100) 的随机数
	num := rand.Int31n(100)
	if num < i.threshold.Load() {
		return i.remote
	}
	return i.local
}

func (i *GrayScaleInteractiveClient) UpdateThreshold(val int32) {
	i.threshold.Store(val)
}
