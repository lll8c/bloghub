package client

import (
	"context"
	intrv1 "geektime/webook/api/proto/gen/intr/v1"
	"geektime/webook/interactive/domain"
	"geektime/webook/interactive/service"
	"google.golang.org/grpc"
)

// InteractiveServiceAdapter
// 将本地调用适配给InteractiveServiceClient客户端
type InteractiveServiceAdapter struct {
	svc service.InteractiveService
}

func NewInteractiveServiceAdapter(svc service.InteractiveService) *InteractiveServiceAdapter {
	return &InteractiveServiceAdapter{svc: svc}
}

func (i InteractiveServiceAdapter) IncrReadCnt(ctx context.Context, in *intrv1.IncrReadCntRequest, opts ...grpc.CallOption) (*intrv1.IncrReadCntResponse, error) {
	err := i.svc.IncrReadCnt(ctx, in.Biz, in.BizId)
	return &intrv1.IncrReadCntResponse{}, err
}

func (i InteractiveServiceAdapter) Like(ctx context.Context, in *intrv1.LikeRequest, opts ...grpc.CallOption) (*intrv1.LikeResponse, error) {
	err := i.svc.Like(ctx, in.Biz, in.BizId, in.Uid)
	return &intrv1.LikeResponse{}, err
}

func (i InteractiveServiceAdapter) CancelLike(ctx context.Context, in *intrv1.CancelLikeRequest, opts ...grpc.CallOption) (*intrv1.CancelLikeResponse, error) {
	err := i.svc.CancelLike(ctx, in.Biz, in.BizId, in.Uid)
	return &intrv1.CancelLikeResponse{}, err
}

func (i InteractiveServiceAdapter) Collect(ctx context.Context, in *intrv1.CollectRequest, opts ...grpc.CallOption) (*intrv1.CollectResponse, error) {
	err := i.svc.Collect(ctx, in.Biz, in.BizId, in.Cid, in.Uid)
	return &intrv1.CollectResponse{}, err
}

func (i InteractiveServiceAdapter) Get(ctx context.Context, in *intrv1.GetRequest, opts ...grpc.CallOption) (*intrv1.GetResponse, error) {
	intr, err := i.svc.Get(ctx, in.Biz, in.BizId, in.Uid)
	return &intrv1.GetResponse{
		Intr: i.toDTO(intr),
	}, err
}

func (i InteractiveServiceAdapter) GetByIds(ctx context.Context, in *intrv1.GetByIdsRequest, opts ...grpc.CallOption) (*intrv1.GetByIdsResponse, error) {
	res, err := i.svc.GetByIds(ctx, in.Biz, in.Ids)
	if err != nil {
		return nil, err
	}
	m := make(map[int64]*intrv1.Interactive, len(res))
	for bizId, intr := range res {
		m[bizId] = i.toDTO(intr)
	}
	return &intrv1.GetByIdsResponse{
		Intrs: m,
	}, nil
}

// DTO data transfer obje0ct
func (i *InteractiveServiceAdapter) toDTO(intr domain.Interactive) *intrv1.Interactive {
	return &intrv1.Interactive{
		Biz:        intr.Biz,
		BizId:      intr.BizId,
		ReadCnt:    intr.ReadCnt,
		LikeCnt:    intr.LikeCnt,
		CollectCnt: intr.CollectCnt,
		Liked:      intr.Liked,
		Collected:  intr.Collected,
	}
}
