package queueworker

import (
	"errors"

	pb "github.com/owlmq/owlmq/api/owlmq"
)

func (s *SimpleQueueWorker) GetStream(consumerUUID string) (pb.Owlmq_ConsumeServer, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if value, exists := s.Consumers[consumerUUID]; exists {
		return value, nil
	} else {
		return nil, errors.New("connection stream not found in map")
	}
}

func (s *SimpleQueueWorker) AddConsumer(consumerUUID string, stream pb.Owlmq_ConsumeServer) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Consumers[consumerUUID] = stream
}

func (s *SimpleQueueWorker) CheckConsumer(consumerUUID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.Consumers[consumerUUID]; exists {
		return nil
	} else {
		return errors.New("consumer not found")
	}
}
