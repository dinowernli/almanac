package discovery

import (
	"fmt"

	pb_almanac "dinowernli.me/almanac/proto"

	"google.golang.org/grpc"
)

// Discovery can be used to find other services in the system.
type Discovery struct {
	appenders []pb_almanac.AppenderClient
}

// New returns a instance of Discovery with clients for the supplied appenders.
func New(appenderEndpoints []string) (*Discovery, error) {
	appenders := []pb_almanac.AppenderClient{}
	for _, endpoint := range appenderEndpoints {
		connection, err := grpc.Dial(endpoint, grpc.WithInsecure())
		if err != nil {
			return nil, fmt.Errorf("unable to dial endpoint %s: %v", endpoint, err)
		}
		appenders = append(appenders, pb_almanac.NewAppenderClient(connection))
	}

	return &Discovery{appenders: appenders}, nil
}

// ListAppenders returns a list of clients, one each per appender in the
// system.
func (d *Discovery) ListAppenders() []pb_almanac.AppenderClient {
	return d.appenders
}
