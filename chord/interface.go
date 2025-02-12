package chord

type Chord_layer interface {
	//get returns value and error if not found
	Get(key string) (value string, err error)
	//put updates a value if already existing and insert it if not
	Put(key string, value string) (err error)

	ShowFingerTable() []string
	GetFingerTable() *FingerTable
	FindSuccessor(joining_node string, request_node string) string
}
