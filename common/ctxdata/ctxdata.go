package ctxdata

import (
	"context"
	"encoding/json"
)

// GetUserIdFromCtx 从 context 中获取登录用户的 ID
func GetUserIdFromCtx(ctx context.Context) uint64 {
	var uid uint64
	if jsonUid, ok := ctx.Value("user_id").(json.Number); ok {
		if int64Uid, err := jsonUid.Int64(); err == nil {
			uid = uint64(int64Uid)
		}
	}
	return uid
}
