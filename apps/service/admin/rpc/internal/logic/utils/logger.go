package utils

import (
	"context"
	"database/sql"
	"encoding/json"

	"github.com/ZY0506/PrimeMall/apps/service/admin/rpc/internal/model"
	"github.com/ZY0506/PrimeMall/apps/service/admin/rpc/internal/svc"
	"github.com/ZY0506/PrimeMall/common/ctxdata"

	"github.com/zeromicro/go-zero/core/logx"
)

// 模块常量
const (
	LogModuleAdmin      = "管理员管理"
	LogModuleUser       = "用户管理"
	LogModuleProduct    = "商品管理"
	LogModuleOrder      = "订单管理"
	LogModuleAfterSale  = "售后管理"
	LogModuleCoupon     = "优惠券管理"
	LogModuleCategory   = "分类管理"
	LogModuleFreight    = "运费模板管理"
	LogModuleRole       = "角色管理"
	LogModulePermission = "权限管理"
)

// RecordAdminLog 记录管理员操作日志
func RecordAdminLog(ctx context.Context, svcCtx *svc.ServiceContext, module string, action string, req interface{}, resp interface{}, durationMs int64) {
	adminId, err := ctxdata.GetUserIdFromCtx(ctx)
	if err != nil {
		logx.WithContext(ctx).Errorf("获取管理员ID失败,err=%v", err)
		return
	}

	// 获取管理员信息
	adminInfo, err := svcCtx.AdminModel.FindOne(ctx, adminId)
	if err != nil {
		logx.WithContext(ctx).Errorf("获取管理员信息失败,adminId=%d,err=%v", adminId, err)
		return
	}

	// 获取客户端信息
	clientInfo, err := ctxdata.GetClientInfoFromCtx(ctx)
	ip := ""
	if err == nil && clientInfo != nil {
		ip = clientInfo.IP
	}

	// 序列化请求和响应
	reqJson := ""
	respJson := ""
	if req != nil {
		if reqBytes, err := json.Marshal(req); err == nil {
			reqJson = string(reqBytes)
		}
	}
	if resp != nil {
		if respBytes, err := json.Marshal(resp); err == nil {
			respJson = string(respBytes)
		}
	}

	// 插入日志
	logx.WithContext(ctx).Infof("记录操作日志,adminId=%d,module=%s,action=%s", adminId, module, action)
	_, err = svcCtx.AdminLogModel.Insert(ctx, &model.AdminLog{
		AdminId:        adminId,
		Username:       adminInfo.Username,
		Module:         module,
		Action:         action,
		RequestMethod:  "RPC",
		RequestUrl:     "",
		RequestParams:  sqlNullString(reqJson),
		ResponseResult: sqlNullString(respJson),
		Ip:             ip,
		DurationMs:     durationMs,
	})
	if err != nil {
		logx.WithContext(ctx).Errorf("记录操作日志失败,adminId=%d,module=%s,err=%v", adminId, module, err)
	}
}

func sqlNullString(s string) sql.NullString {
	return sql.NullString{
		String: s,
		Valid:  s != "",
	}
}
