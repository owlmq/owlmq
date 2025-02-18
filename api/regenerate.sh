#!/bin/bash

#owlmq API
rm ./owlmq/owlmq_grpc.pb.go
rm ./owlmq/owlmq.pb.go
protoc --go_out=. --go-grpc_out=. ./owlmq/owlmq.proto


#plugin API
rm ./plugin/plugin_grpc.pb.go
rm ./plugin/plugin.pb.go
protoc --go_out=. --go-grpc_out=. ./plugin/plugin.proto
