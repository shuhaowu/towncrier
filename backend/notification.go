package backend

type Priority int64

const (
	LowPriority    Priority = 1   // reserved for future use
	NormalPriority Priority = 2   // regular stuff
	UrgentPriority Priority = 100 // notify now
)

var PriorityMap map[string]Priority = map[string]Priority{
	"low":    LowPriority,
	"normal": NormalPriority,
	"urgent": UrgentPriority,
}

type Notification struct {
	Subject   string
	Content   string
	Channel   string
	Origin    string
	Tags      []string
	Priority  Priority
	CreatedAt int64 // UnixNano
	UpdatedAt int64 // UnixNano
}
