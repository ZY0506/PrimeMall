package userlogic

import (
	"context"
	"time"

	"github.com/ZY0506/PrimeMall/common/ctxdata"
	"github.com/ZY0506/PrimeMall/pkg/oss"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/ZY0506/PrimeMall/apps/service/user/rpc/internal/svc"
	"github.com/ZY0506/PrimeMall/apps/service/user/rpc/types/user"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetOssTokenLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetOssTokenLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetOssTokenLogic {
	return &GetOssTokenLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// GetOssToken 获取OSS上传凭证
func (l *GetOssTokenLogic) GetOssToken(in *user.OssTokenReq) (*user.OssTokenResp, error) {
	// 获取当前账号信息
	userId, err := ctxdata.GetUserIdFromCtx(l.ctx)
	if err != nil {
		l.Logger.Errorf("获取用户ID失败,error=%v", err)
		return nil, err
	}
	token, err := oss.GeneratePolicyToken(oss.OssConfig{
		AccessKeyId:     l.svcCtx.Config.OSSConfig.AccessKeyId,
		AccessKeySecret: l.svcCtx.Config.OSSConfig.AccessKeySecret,
		BucketName:      l.svcCtx.Config.OSSConfig.BucketName,
		Host:            l.svcCtx.Config.OSSConfig.Host,
		UploadDir:       l.svcCtx.Config.OSSConfig.UploadDir,
		ExpireTime:      l.svcCtx.Config.OSSConfig.ExpireTime,
		CallbackUrl:     l.svcCtx.Config.OSSConfig.CallbackUrl,
		MaxFileSize:     l.svcCtx.Config.OSSConfig.MaxFileSize,
		AllowedExts:     l.svcCtx.Config.OSSConfig.AllowedExts,
	}, in.FileName, userId)
	if err != nil {
		l.Logger.Errorf("生成OSS上传凭证失败,error=%v", err)
		return nil, err
	}

	l.Logger.Info("生成OSS上传凭证成功")
	return &user.OssTokenResp{
		AccessId:  token.AccessKeyId,
		Host:      token.Host,
		Policy:    token.Policy,
		Signature: token.Signature,
		Expire:    timestamppb.New(time.Unix(token.ExpireTimestamp, 0)),
		Dir:       token.Dir,
		ObjectKey: token.ObjectKey,
		Callback:  token.Callback,
	}, nil
}
