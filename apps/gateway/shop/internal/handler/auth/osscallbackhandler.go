// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.1

package auth

import (
	"github.com/ZY0506/PrimeMall/apps/gateway/shop/internal/logic/auth"
	"github.com/ZY0506/PrimeMall/apps/gateway/shop/internal/svc"
	"github.com/ZY0506/PrimeMall/apps/gateway/shop/internal/types"
	"github.com/zeromicro/go-zero/core/logx"
	"net/http"
	"strconv"
)

// OssCallbackHandler OSS上传回调（服务端自行幂等）
func OssCallbackHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 1. 限制请求体大小（防止恶意请求）
		r.Body = http.MaxBytesReader(w, r.Body, 1<<20)

		// 2. 只允许 POST
		if r.Method != http.MethodPost {
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
			return
		}

		// 4. 解析 form
		if err := r.ParseForm(); err != nil {
			http.Error(w, "Bad Request", http.StatusBadRequest)
			return
		}

		// 5. 获取 OSS 回调字段
		sizeStr := r.FormValue("size")

		size, err := strconv.ParseInt(sizeStr, 10, 64)
		if err != nil {
			http.Error(w, "Invalid size", http.StatusBadRequest)
			return
		}

		req := &types.OssCallbackReq{
			Bucket: r.FormValue("bucket"),
			Object: r.FormValue("object"),
			Etag:   r.FormValue("etag"),
			Size:   size,
		}

		// 6. 基础校验
		if req.Object == "" {
			http.Error(w, "Bad Request", http.StatusBadRequest)
			return
		}

		// 7. 业务处理
		l := auth.NewOssCallbackLogic(r.Context(), svcCtx)

		if err := l.OssCallback(req); err != nil {
			logx.WithContext(r.Context()).Errorf("OSS callback business failed: %v", err)
			// OSS 会自动重试
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		// 8. 成功返回
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	}
}
