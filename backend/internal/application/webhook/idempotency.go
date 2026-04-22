package webhook

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	"github.com/yuanjun5681/clawhire/backend/internal/protocol/clawsynapse"
)

// DeriveEventKey 按如下优先级生成幂等键（backend_design.md §八）：
//
//  1. metadata.eventId
//  2. "cs:" + sessionKey + ":" + type + ":" + (metadata.taskId | metadata.businessId)
//  3. "h:" + sha256(type + "|" + sessionKey + "|" + message)
//
// 任一级别能确定唯一键就返回，保证同一事件多次投递得到相同 eventKey。
func DeriveEventKey(env *clawsynapse.Envelope) string {
	if env == nil {
		return ""
	}
	if v := env.MetaString("eventId"); v != "" {
		return v
	}

	bizID := env.MetaString("taskId")
	if bizID == "" {
		bizID = env.MetaString("businessId")
	}
	if bizID != "" && env.SessionKey != "" && env.Type != "" {
		return fmt.Sprintf("cs:%s:%s:%s", env.SessionKey, env.Type, bizID)
	}

	h := sha256.Sum256([]byte(env.Type + "|" + env.SessionKey + "|" + env.Message))
	return "h:" + hex.EncodeToString(h[:])
}
