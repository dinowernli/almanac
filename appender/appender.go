package appender

import (
	pb_almanac "dinowernli.me/almanac/proto"
)

type appender struct {
}

func New() *appender {
	return &appender{}
}

func (a *appender) Append(request *pb_almanac.AppendRequest) {
	// TODO(dino): Change signature to implement grpc service.
}
