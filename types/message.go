package types

type Queue struct {
	Key   string     `json:"key"` //-> "q:logs"
	Value QueueValue `json:"value"`
}

type ConsumerEntry struct {
	ConsumerUUID      string `json:"consumer_uuid"`
	QueueWorkerAdress string `json:"queue_worker_adress"`
}

type QueueValue struct {
	Consumers []ConsumerEntry `json:"consumers"`
	HeadUUID  string          `json:"headuuid"`
	//TODO add this later(storage layer need to support this)
	Persistent bool `json:"persistent"`
}

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
