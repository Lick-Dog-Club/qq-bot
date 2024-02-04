package types

type Role string

func (t Role) String() string {
	return string(t)
}

const (
	RoleUser      Role = "user"
	RoleSystem    Role = "system"
	RoleAssistant Role = "assistant"
)
