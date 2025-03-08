package types

type Message struct {
	Key   string //-> "m:uuid"
	Value MessageValue
}

type MessageValue struct {
	UUID string //-> e.g.: "12345"
	//TODO maybe refactor to be the routing key
	QueueName      string //-> e.g.: "logs"
	Content        string
	PreMessageUUID string //-> e.g.: "12345"
	//genesis is the first msg in a queue
	IsGenesis bool
}
