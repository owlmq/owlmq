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
	//Persistent bool `json:"persistent"`
}
