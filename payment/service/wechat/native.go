package wechat

import (
	"context"
	"errors"
	"fmt"
	"geektime/webook/payment/domain"
	"geektime/webook/payment/events"
	"geektime/webook/payment/repository"
	"geektime/webook/pkg/logger"
	"github.com/wechatpay-apiv3/wechatpay-go/core"
	"github.com/wechatpay-apiv3/wechatpay-go/services/payments"
	"github.com/wechatpay-apiv3/wechatpay-go/services/payments/native"
	"time"
)

var errUnknownTransactionState = errors.New("未知的微信事务状态")

type NativePaymentService struct {
	//公众号id
	appID string
	//商户号
	mchID string
	// 支付通知回调 URL
	notifyURL string
	// 自己的支付记录
	repo repository.PaymentRepository

	svc      *native.NativeApiService
	producer events.Producer

	l logger.LoggerV1

	// 在微信 native 里面，分别是
	// SUCCESS：支付成功
	// REFUND：转入退款
	// NOTPAY：未支付
	// CLOSED：已关闭
	// REVOKED：已撤销（付款码支付）
	// USERPAYING：用户支付中（付款码支付）
	// PAYERROR：支付失败(其他原因，如银行返回失败)
	nativeCBTypeToStatus map[string]domain.PaymentStatus
}

func NewNativePaymentService(appID string, mchID string,
	repo repository.PaymentRepository, svc *native.NativeApiService,
	l logger.LoggerV1, producer events.Producer) *NativePaymentService {
	return &NativePaymentService{appID: appID, mchID: mchID, notifyURL: "http://wechat.meoying.com/pay/callback",
		repo: repo, svc: svc, l: l, producer: producer,
		nativeCBTypeToStatus: map[string]domain.PaymentStatus{
			"SUCCESS":  domain.PaymentStatusSuccess,
			"PAYERROR": domain.PaymentStatusFailed,
			"NOTPAY":   domain.PaymentStatusInit,
			"CLOSED":   domain.PaymentStatusFailed,
			"REVOKED":  domain.PaymentStatusFailed,
			"REFUND":   domain.PaymentStatusRefund,
			// 其它状态你都可以加
		},
	}
}

// Prepay 生成订单拿到扫码支付的二维码
func (n *NativePaymentService) Prepay(ctx context.Context, pmt domain.Payment) (string, error) {
	//设置支付订单状态为初始态
	pmt.Status = domain.PaymentStatusInit
	//创建支付订单
	err := n.repo.AddPayment(ctx, pmt)
	if err != nil {
		return "", err
	}
	//返回的result有http请求和响应
	resp, _, err := n.svc.Prepay(ctx, native.PrepayRequest{
		Appid:       core.String(n.appID),
		Mchid:       core.String(n.mchID),
		Description: core.String(pmt.Description),
		//业务方一定要传一个唯一标识去重
		OutTradeNo: core.String(pmt.BizTradeNO),
		// 最好这个要带上
		TimeExpire: core.Time(time.Now().Add(time.Minute * 30)),
		NotifyUrl:  core.String(n.notifyURL),
		Amount: &native.Amount{
			Total:    core.Int64(pmt.Amt.Total),
			Currency: core.String(pmt.Amt.Currency),
		},
	})
	//n.l.Debug("微信prepay响应",
	//	logger.Field{Key: "result", Val: result},
	//	logger.Field{Key: "resp", Val: resp})
	if err != nil {
		return "", err
	}
	//获得二维码链接
	return *resp.CodeUrl, nil
}

func (n *NativePaymentService) GetPayment(ctx context.Context, bizTradeId string) (domain.Payment, error) {
	return n.repo.GetPayment(ctx, bizTradeId)
}

// HandleCallback 处理回调信息
func (n *NativePaymentService) HandleCallback(ctx context.Context, txn *payments.Transaction) error {
	//更新支付记录
	return n.updateByTxn(ctx, txn)
}

// 更新订单状态
func (n *NativePaymentService) updateByTxn(ctx context.Context, txn *payments.Transaction) error {
	//根据transaction中的状态映射成功与失败等状态
	status, ok := n.nativeCBTypeToStatus[*txn.TradeState]
	if !ok {
		return fmt.Errorf("%w, 微信的状态是 %s", errUnknownTransactionState, *txn.TradeState)
	}
	// 很显然，就是更新一下我们本地数据库里面 payment 的状态
	err := n.repo.UpdatePayment(ctx, domain.Payment{
		// 微信过来的 transaction id
		TxnID:      *txn.TransactionId,
		BizTradeNO: *txn.OutTradeNo,
		Status:     status,
	})
	if err != nil {
		return err
	}
	// 就要通知业务方了
	// 有些人的系统，会根据支付状态来决定要不要通知
	// 我要是发消息失败了怎么办？
	// 站在业务的角度，你是不是至少应该发成功一次
	err1 := n.producer.ProducePaymentEvent(ctx, events.PaymentEvent{
		BizTradeNO: *txn.OutTradeNo,
		Status:     status.AsUint8(),
	})
	if err1 != nil {
		n.l.Error("发送支付事件失败", logger.Error(err),
			logger.String("biz_trade_no", *txn.OutTradeNo))
	}
	return nil
}

// SyncWechatInfo 同步微信订单状态
func (n *NativePaymentService) SyncWechatInfo(ctx context.Context, bizTradeNO string) error {
	// 对账
	//主动查询订单状态
	txn, _, err := n.svc.QueryOrderByOutTradeNo(ctx, native.QueryOrderByOutTradeNoRequest{
		OutTradeNo: core.String(bizTradeNO),
		Mchid:      core.String(n.mchID),
	})
	if err != nil {
		return err
	}
	//更新订单
	return n.updateByTxn(ctx, txn)
}

// FindExpiredPayment 查找过期的订单
func (n *NativePaymentService) FindExpiredPayment(ctx context.Context, offset, limit int, t time.Time) ([]domain.Payment, error) {
	return n.repo.FindExpiredPayment(ctx, offset, limit, t)
}
