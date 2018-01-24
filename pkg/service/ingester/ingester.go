package ingester

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	almHttp "dinowernli.me/almanac/pkg/http"
	dc "dinowernli.me/almanac/pkg/service/discovery"
	"dinowernli.me/almanac/pkg/util"
	pb_almanac "dinowernli.me/almanac/proto"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
)

const (
	httpUrl             = "/ingester"
	httpIngestTimeoutMs = 300
	urlParamContent     = "c"

	httpExampleEntry = `
{
	"message": "foo",
	"timestamp_ms": 12345,
	"logger": "MyAwesomeLogger"
}
`

	timestampField = "timestamp_ms"
	nanosPerMilli  = 1000000
)

var (
	ingestField = logrus.Fields{"method": "ingester.Ingest"}
)

// Ingester is an implementation of the ingester service. It accepts log
// entries entering the system and fans them out to appenders.
type Ingester struct {
	logger       *logrus.Logger
	discovery    *dc.Discovery
	ingestFanout int
}

// New returns a new Ingester backed by the supplied service discovery.
// appenderFanout specifies how many appenders this ingester tries to inform of
// a new log entry before declaring the entry ingested into the system.
func New(logger *logrus.Logger, discovery *dc.Discovery, ingestFanout int) (*Ingester, error) {
	if ingestFanout < 1 {
		return nil, fmt.Errorf("ingestFanout must be at least 1")
	}

	return &Ingester{
		logger:       logger,
		discovery:    discovery,
		ingestFanout: ingestFanout,
	}, nil
}

// RegisterHttp registers a page on the supplied server, used for ingesting entries.
func (i *Ingester) RegisterHttp(server *http.ServeMux) {
	server.HandleFunc(httpUrl, prometheus.InstrumentHandlerFunc(httpUrl, i.handleHttp))
}

func (i *Ingester) Ingest(ctx context.Context, request *pb_almanac.IngestRequest) (*pb_almanac.IngestResponse, error) {
	logger := i.logger.WithFields(ingestField)

	// Parse the incoming raw log entry, extracting some structure.
	entry, err := extractEntry(request.EntryJson)
	if err != nil {
		err := grpc.Errorf(codes.InvalidArgument, "unable to extract log entry from json: %v", err)
		i.logger.WithError(err).Warnf("Failed")
		return nil, err
	}
	logger = logger.WithFields(logrus.Fields{"entry": entry.Id})

	// Send an append request to a select bunch of appenders.
	appenders, err := i.selectAppenders()
	if err != nil {
		err := grpc.Errorf(codes.Internal, "unable to select appenders: %v", err)
		logger.WithError(err).Warnf("Failed")
		return nil, err
	}

	appendRequest := &pb_almanac.AppendRequest{Entry: entry}
	resultChan := make(chan error, len(appenders))
	for _, appender := range appenders {
		go func() {
			_, err := appender.Append(ctx, appendRequest)
			resultChan <- err
		}()
	}

	// TODO(dino): be smarter about how many append calls need to have succeeded. For now,
	// we error out if any of the append calls fail.
	for i := 0; i < len(appenders); i++ {
		err := <-resultChan
		if err != nil {
			err := grpc.Errorf(codes.Internal, "unable to send append request: %v", err)
			logger.WithError(err).Warnf("Failed")
			return nil, err
		}
	}

	logger.Infof("Handled")
	return &pb_almanac.IngestResponse{}, nil
}

// selectAppenders returns the appenders to whom the current log entry must be
// sent to.
func (i *Ingester) selectAppenders() ([]pb_almanac.AppenderClient, error) {
	allAppenders := i.discovery.ListAppenders()
	if i.ingestFanout > len(allAppenders) {
		return nil, fmt.Errorf("cannot select %d appenders from a list of size %d", i.ingestFanout, len(allAppenders))
	}

	// TODO(dino): Consider remembering the appenders last used and trying to find them.
	// Shuffle the first time so that different ingesters use different subsets of appenders.

	return allAppenders[0:i.ingestFanout], nil
}

// handleHttp serves a web page which can be used to ingest entries on this ingester.
func (i *Ingester) handleHttp(writer http.ResponseWriter, request *http.Request) {
	pageData := &almHttp.IngesterData{FormContent: request.FormValue(urlParamContent)}
	if pageData.FormContent != "" {
		ctx, cancel := context.WithTimeout(context.Background(), httpIngestTimeoutMs*time.Millisecond)
		defer cancel()

		_, pageData.Error = i.Ingest(ctx, &pb_almanac.IngestRequest{EntryJson: pageData.FormContent})
		if pageData.Error == nil {
			pageData.Result = "Successfully ingested entry"
		}
	} else {
		// Populate the form with some example data.
		pageData.FormContent = httpExampleEntry
	}

	err := pageData.Render(writer)
	if err != nil {
		fmt.Fprintf(writer, "failed to render ingester page: %v", err)
	}
}

// extractEntry takes an incoming string and construct a LogEntry proto from
// it. This can fail if the incoming entry is not valid json, or if the json
// is otherwise malformed.
func extractEntry(rawJson string) (*pb_almanac.LogEntry, error) {
	var rawEntry map[string]*json.RawMessage
	err := json.Unmarshal([]byte(rawJson), &rawEntry)
	if err != nil {
		return nil, fmt.Errorf("unable to parse raw entry")
	}

	timestampMs := int64(0)

	// Attempt to extract a timestamp.
	value, ok := rawEntry[timestampField]
	if ok {
		err := json.Unmarshal(*value, &timestampMs)
		if err != nil {
			return nil, fmt.Errorf("could not parse value for timestamp: %v", err)
		}
	}

	// If no timestamp was provided, just use "now".
	if timestampMs == 0 {
		timestampMs = time.Now().UnixNano() / nanosPerMilli
	}

	return &pb_almanac.LogEntry{
		EntryJson:   rawJson,
		TimestampMs: timestampMs,
		Id:          newEntryId(timestampMs),
	}, nil
}

// newEntryId returns a string id for an entry with the given timestamp. The
// ids have the property that sorting them lexicographically orders them by
// timestamp, but that no two different entries ever end up with the same id.
func newEntryId(timestampMs int64) string {
	return fmt.Sprintf("%d-%s", timestampMs, util.RandomString(3))
}
