package replicator

type ReplicatorType int

const (
	SuccessorList ReplicatorType = iota
	VirtualNode
)

func (r ReplicatorType) String() string {
	switch r {
	case SuccessorList:
		return "SuccessorList"
	case VirtualNode:
		return "VirtualNode"
	default:
		return "Unknown"
	}
}
