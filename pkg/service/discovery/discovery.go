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

// New returns an instance which talks to appenders at the supplied addresses over grpc.
func New(appenderEndpoints []string) (*Discovery, error) {
	appenders := []pb_almanac.AppenderClient{}
	for _, endpoint := range appenderEndpoints {
		connection, err := grpc.Dial(endpoint, grpc.WithInsecure())
		if err != nil {
			return nil, fmt.Errorf("unable to dial endpoint %s: %v", endpoint, err)
		}
		appenders = append(appenders, pb_almanac.NewAppenderClient(connection))
	}

	return &Discovery{appenders}, nil
}

// NewForTesting resturns an instance which talks directly to the supplied appenders.
// This should only be used for testing.
func NewForTesting(appenders []pb_almanac.AppenderClient) *Discovery {
	return &Discovery{appenders}
}

// ListAppenders returns a list of clients, one each per appender in the
// system. The returned list if a snapshot of the discovery object's
// canonical list, so callers my modify the returned list.
func (d *Discovery) ListAppenders() []pb_almanac.AppenderClient {
	result := []pb_almanac.AppenderClient{}
	for _, a := range d.appenders {
		result = append(result, a)
	}
	return result
}
