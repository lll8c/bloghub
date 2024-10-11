package grpc

import (
	"context"
	intrv1 "geektime/webook/api/proto/gen/intr/v1"
	"geektime/webook/interactive/domain"
	"geektime/webook/interactive/service"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// InteractiveServiceServer 这里只是把service包装成一个grpc
// 和grpc有关的操作限定在这里
type InteractiveServiceServer struct {
	intrv1.UnimplementedInteractiveServiceServer
	//封装好业务的核心逻辑
	svc service.InteractiveService
}

func NewInteractiveServiceServer(svc service.InteractiveService) *InteractiveServiceServer {
	return &InteractiveServiceServer{svc: svc}
}

func (i *InteractiveServiceServer) Register(server *grpc.Server) {
	intrv1.RegisterInteractiveServiceServer(server, i)
}

func (i *InteractiveServiceServer) IncrReadCnt(ctx context.Context, request *intrv1.IncrReadCntRequest) (*intrv1.IncrReadCntResponse, error) {
	err := i.svc.IncrReadCnt(ctx, request.Biz, request.BizId)
	return &intrv1.IncrReadCntResponse{}, err
}

func (i InteractiveServiceServer) Like(ctx context.Context, request *intrv1.LikeRequest) (*intrv1.LikeResponse, error) {
	err := i.svc.Like(ctx, request.Biz, request.BizId, request.Uid)
	return &intrv1.LikeResponse{}, err
}

func (i InteractiveServiceServer) CancelLike(ctx context.Context, request *intrv1.CancelLikeRequest) (*intrv1.CancelLikeResponse, error) {
	//可以做参数校验
	//也可以用grpc的插件
	if request.Uid <= 0 {
		return nil, status.Error(codes.InvalidArgument, "uid 错误")
	}
	err := i.svc.CancelLike(ctx, request.Biz, request.BizId, request.Uid)
	return &intrv1.CancelLikeResponse{}, err
}

func (i InteractiveServiceServer) Collect(ctx context.Context, request *intrv1.CollectRequest) (*intrv1.CollectResponse, error) {
	err := i.svc.Collect(ctx, request.Biz, request.BizId, request.Cid, request.Uid)
	return &intrv1.CollectResponse{}, err
}

func (i InteractiveServiceServer) Get(ctx context.Context, request *intrv1.GetRequest) (*intrv1.GetResponse, error) {
	res, err := i.svc.Get(ctx, request.Biz, request.BizId, request.Uid)
	if err != nil {
		return nil, err
	}
	return &intrv1.GetResponse{
		Intr: i.toDTO(res),
	}, nil
}

func (i InteractiveServiceServer) GetByIds(ctx context.Context, request *intrv1.GetByIdsRequest) (*intrv1.GetByIdsResponse, error) {
	res, err := i.svc.GetByIds(ctx, request.Biz, request.Ids)
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

// DTO data transfer object
func (i *InteractiveServiceServer) toDTO(intr domain.Interactive) *intrv1.Interactive {
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
