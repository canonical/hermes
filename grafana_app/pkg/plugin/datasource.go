package plugin

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/grafana/grafana-plugin-sdk-go/backend/httpclient"
	"github.com/grafana/grafana-plugin-sdk-go/backend/instancemgmt"
	"github.com/grafana/grafana-plugin-sdk-go/backend/log"
)

var (
	_ backend.QueryDataHandler      = (*Datasource)(nil)
	_ backend.CheckHealthHandler    = (*Datasource)(nil)
	_ instancemgmt.InstanceDisposer = (*Datasource)(nil)
)

func NewDatasource(settings backend.DataSourceInstanceSettings) (instancemgmt.Instance, error) {
	opts, err := settings.HTTPClientOptions()
	if err != nil {
		return nil, fmt.Errorf("http client options %w", err)
	}
	client, err := httpclient.New(opts)
	if err != nil {
		return nil, fmt.Errorf("http client new %w", err)
	}
	return &Datasource{
		settings:   settings,
		httpClient: client,
	}, nil
}

type Datasource struct {
	settings   backend.DataSourceInstanceSettings
	httpClient *http.Client
}

func (ds *Datasource) Dispose() {
	ds.httpClient.CloseIdleConnections()
}

func (ds *Datasource) QueryData(ctx context.Context, req *backend.QueryDataRequest) (*backend.QueryDataResponse, error) {
	response := backend.NewQueryDataResponse()

	for _, q := range req.Queries {
		var err error
		res := backend.DataResponse{}
		defer func() {
			res.Error = err
			response.Responses[q.RefID] = res
		}()
		if query, err := ParseQuery(q); err != nil {
			log.DefaultLogger.Error("query data: failed to parse query", "err", err)
		} else if urlPath, err := url.JoinPath(ds.settings.URL, query.Group, query.Routine); err != nil {
			log.DefaultLogger.Error("query data: failed to join path", "err", err)
		} else {
			interval := int64(q.Interval.Seconds())
			startTime := q.TimeRange.From.Unix() - interval
			endTime := q.TimeRange.To.Unix() + interval
			if resp, err := ds.query(ctx, urlPath); err != nil {
				log.DefaultLogger.Error("query data: failed to query", "err", err)
			} else if frames, err := HandleQueryResp(query.Group, query.Routine, startTime, endTime, resp); err != nil {
				log.DefaultLogger.Error("query data: failed to handle resp", "err", err)
			} else {
				res.Frames = append(res.Frames, frames...)
			}
		}
	}
	return response, nil
}

func (ds *Datasource) query(ctx context.Context, url string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	httpResp, err := ds.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := httpResp.Body.Close(); err != nil {
			log.DefaultLogger.Error("query: failed to close response body", "err", err)
		}
	}()

	if httpResp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("bad http status %d", httpResp.StatusCode)
	}

	response, err := ioutil.ReadAll(httpResp.Body)
	if err != nil {
		return nil, err
	}
	return response, nil
}

func newHealthCheckErrorf(format string, args ...interface{}) *backend.CheckHealthResult {
	return &backend.CheckHealthResult{Status: backend.HealthStatusError, Message: fmt.Sprintf(format, args...)}
}

func (ds *Datasource) CheckHealth(ctx context.Context, _ *backend.CheckHealthRequest) (*backend.CheckHealthResult, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, ds.settings.URL, nil)
	if err != nil {
		return newHealthCheckErrorf("could not create request"), nil
	}
	resp, err := ds.httpClient.Do(req)
	if err != nil {
		return newHealthCheckErrorf("request error"), nil
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			log.DefaultLogger.Error("check health: failed to close response body", "err", err.Error())
		}
	}()

	if resp.StatusCode != http.StatusOK {
		return newHealthCheckErrorf("got response code %d", resp.StatusCode), nil
	}
	return &backend.CheckHealthResult{
		Status:  backend.HealthStatusOk,
		Message: "Data source is working",
	}, nil
}

func (ds *Datasource) CallResource(ctx context.Context, req *backend.CallResourceRequest,
	sender backend.CallResourceResponseSender) error {
	status := http.StatusInternalServerError
	respBody := []byte{}

	defer func() {
		sender.Send(&backend.CallResourceResponse{
			Status: status,
			Body:   respBody,
		})
	}()

	_url, err := url.Parse(req.URL)
	if err != nil {
		return err
	}

	urlPath, err := url.JoinPath(ds.settings.URL, _url.Path)
	if err != nil {
		return err
	}

	resp, err := ds.query(ctx, urlPath)
	if err != nil {
		log.DefaultLogger.Error("query res: failed to query", "err", err)
		return err
	}
	return HandleResResp(resp, &status, &respBody)
}
