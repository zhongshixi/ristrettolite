package emission

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type Scope3API struct {

	// https://api.scope3.com/v2/measure?includeRows=true&latest=true&fields=emissionsBreakdown
	endpoint        string
	measurementPath string
	token           string

	Client *http.Client
}

func NewMeasurementAPI(token string) *Scope3API {
	return &Scope3API{
		endpoint:        "https://api.scope3.com/v2",
		measurementPath: "/measure?includeRows=true&latest=true&fields=emissionsBreakdown",
		token:           token,
		Client: &http.Client{
			Transport: &http.Transport{},
		},
	}
}

func (m *Scope3API) GetEmissionMeasurement(ctx context.Context, payload EmissionRequestPayload) (*EmissionResponsePayload, error) {
	payloadBy, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", m.endpoint+m.measurementPath, bytes.NewBuffer(payloadBy))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Add("accept", "application/json")
	req.Header.Add("content-type", "application/json")
	req.Header.Add("authorization", "Bearer "+m.token)

	res, err := m.Client.Do(req)
	defer func() {
		if res != nil {
			res.Body.Close()
		}
	}()

	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("request failed with status code %d", res.StatusCode)
	}

	resPayloadBytes, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	data := &EmissionResponsePayload{}
	if err := json.Unmarshal(resPayloadBytes, data); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response body: %w", err)
	}

	return data, nil
}
