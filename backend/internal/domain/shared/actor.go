package shared

type ActorKind string

const (
	ActorKindUser   ActorKind = "user"
	ActorKindAgent  ActorKind = "agent"
	ActorKindSystem ActorKind = "system"
)

// Actor 描述需求方 / 执行方 / 验收方 / 收款方等角色。
// 字段命名与 api_design.md 对外保持一致。
type Actor struct {
	ID   string    `bson:"id"             json:"id"`
	Kind ActorKind `bson:"kind"           json:"kind"`
	Name string    `bson:"name,omitempty" json:"name,omitempty"`
}

func (k ActorKind) Valid() bool {
	switch k {
	case ActorKindUser, ActorKindAgent, ActorKindSystem:
		return true
	}
	return false
}
