package webhook

import (
	"context"

	"github.com/yuanjun5681/clawhire/backend/internal/domain/event"
	"github.com/yuanjun5681/clawhire/backend/internal/protocol/clawsynapse"
)

// Dispatcher 将 clawhire.* 消息路由到具体 Command Handler。
//
// Step 6 阶段我们只打通 Webhook 入站链路，尚未接入业务处理器，
// 因此使用 NoopDispatcher —— 统一返回 ProcessStatusSkipped。
// Step 7 会引入具体的 CommandDispatcher，实现 12 个消息的处理。
type Dispatcher interface {
	Dispatch(ctx context.Context, env *clawsynapse.Envelope) (event.ProcessStatus, error)
}

// NoopDispatcher 只记录"已接收但未处理"，用于 Step 6 的链路验证。
type NoopDispatcher struct{}

func (NoopDispatcher) Dispatch(_ context.Context, _ *clawsynapse.Envelope) (event.ProcessStatus, error) {
	return event.ProcessStatusSkipped, nil
}
