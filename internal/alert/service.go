package alert

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/cicconee/weather-app/internal/geometry"
	"github.com/cicconee/weather-app/internal/nws"
)

type Service struct {
	Client *nws.Client
	Store  *Store
}

func New(client *nws.Client, db *sql.DB) *Service {
	return &Service{
		Client: client,
		Store:  NewStore(db),
	}
}

// Sync fetches and stores all the active alerts for
// each state stored in the database. Any referenced
// alerts will be deleted from the database and the
// most up to date alert will be stored.
//
// A SyncResult is returned stating what states are
// being synced, the total alerts written, and if
// any failures happened while syncing.
func (s *Service) Sync(ctx context.Context) (SyncResult, error) {
	states, err := s.Store.SelectStates(ctx)
	if err != nil {
		return SyncResult{}, fmt.Errorf("failed to select states: %w", err)
	}

	return s.sync(ctx, states)
}

// SyncResult defines the result of syncing
// alerts. It is returned by Sync.
type SyncResult struct {
	States      []State
	TotalWrites int
	Fails       []SyncResourceFail
}

// Fail appends a SyncResourceFail to
// this sync result Fails field.
func (s *SyncResult) Fail(f SyncResourceFail) {
	s.Fails = append(s.Fails, f)
}

// SyncResourceFail defines a failure
// that occured while syncing alerts.
type SyncResourceFail struct {
	// The identifier of the alert that
	// failed.
	ID string

	// The operation that failed.
	Op string

	// The error that caused the failure.
	Err error
}

func (s *Service) sync(ctx context.Context, states []State) (SyncResult, error) {
	result := SyncResult{
		States:      states,
		TotalWrites: 0,
		Fails:       []SyncResourceFail{},
	}

	alerts, err := s.alerts(ctx, states)
	if err != nil {
		return SyncResult{}, fmt.Errorf("failed to fetch active alerts: %w", err)
	}

	for _, a := range alerts {
		_, err := s.Store.SelectAlert(ctx, a.Alert.ID)
		switch {
		case errors.Is(err, sql.ErrNoRows):
			s.write(ctx, a, &result)
		case err != nil:
			result.Fail(SyncResourceFail{ID: a.Alert.ID, Op: "select", Err: err})
		default:
			// Alert already exists in database.
			// Do nothing.
		}
	}

	return result, nil
}

func (s *Service) write(ctx context.Context, e Resource, sync *SyncResult) {
	if err := s.Store.InsertAlertTx(ctx, e); err != nil {
		sync.Fail(SyncResourceFail{ID: e.Alert.ID, Op: "insert", Err: err})
	} else {
		sync.TotalWrites++
	}
}

func (s *Service) alerts(ctx context.Context, states StateCollection) ([]Resource, error) {
	alerts, err := s.Client.GetActiveAlerts(states.AsStrings()...)
	var statusError *nws.StatusCodeError
	switch {
	case err == nil:
		return resourcesFromNWS(alerts), nil
	case errors.As(err, &statusError):
		if statusError.StatusCode == 400 || statusError.StatusCode == 500 {
			return nil, &Error{
				error:      fmt.Errorf("active alerts unreachable: %w", err),
				msg:        "unable to get active alerts",
				statusCode: http.StatusServiceUnavailable,
			}
		}

		return nil, fmt.Errorf("unexpected status code: %w", err)
	default:
		return nil, err
	}
}

// GetResponse is a collection of alerts
// for a specific lon-lat coordinates.
// It is returned by Get.
type GetResponse struct {
	Lon    float64         `json:"lon"`
	Lat    float64         `json:"lat"`
	Alerts AlertCollection `json:"alerts"`
}

// Get gets all the active alerts for point
// and returns it as a GetResponse.
func (s *Service) Get(ctx context.Context, point geometry.Point) (GetResponse, error) {
	collection, err := s.Store.SelectAlertsContains(ctx, point)
	if err != nil {
		return GetResponse{}, err
	}

	return GetResponse{
		Lon:    point.Lon(),
		Lat:    point.Lat(),
		Alerts: collection,
	}, nil
}

// CleanUp will delete any alerts from the database
// that are expired or ended at the time of calling
// this func. It will return the number of rows deleted.
//
// If an error is returned it is still possible that
// some rows were deleted.
func (s *Service) CleanUp(ctx context.Context) (int64, error) {
	t := time.Now().UTC()

	n1, err := s.Store.DeleteEndedAlerts(ctx, t)
	if err != nil {
		return 0, fmt.Errorf("failed to delete alerts with outdated ends time: %w", err)
	}

	n2, err := s.Store.DeleteExpiredAlerts(ctx, t)
	if err != nil {
		return n1, fmt.Errorf("failed to delete alerts with outdated expires time: %w", err)
	}

	return n1 + n2, nil
}

func resourcesFromNWS(alerts []nws.Alert) []Resource {
	e := []Resource{}
	for _, a := range alerts {
		e = append(e, resourceFromNWS(a))
	}
	return e
}

func resourceFromNWS(a nws.Alert) Resource {
	onset := a.OnSet.UTC()
	ends := a.Ends.UTC()

	return Resource{
		Alert: &Alert{
			ID:          a.ID,
			AreaDesc:    a.AreaDesc,
			OnSet:       &onset,
			Ends:        &ends,
			Category:    a.Category,
			Severity:    a.Severity,
			Certainty:   a.Certainty,
			Urgency:     a.Urgency,
			Event:       a.Event,
			Headline:    a.Headline,
			Description: a.Description,
			Instruction: a.Instruction,
			Response:    a.Response,
			Expires:     a.Expires,
			MessageType: a.MessageType,
			Points:      a.Geometry,
		},
		References: referenceCollectionFromNWS(a.References),
		Zones:      zonesFromNWS(a.AffectedZones),
	}
}

func referenceCollectionFromNWS(nwsRefs []nws.AlertReference) ReferenceCollection {
	refs := []Reference{}
	for _, nr := range nwsRefs {
		refs = append(refs, Reference(nr.ID))
	}
	return refs
}

func zonesFromNWS(affected []string) []Zone {
	zones := []Zone{}
	for _, uri := range affected {
		zones = append(zones, Zone{
			URI: uri,
		})
	}
	return zones
}
