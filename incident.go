package victorops

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"time"
)

// Transition represents a state changes of an incident
type Transition struct {
	Name     string    `json:",omitempty"`
	At       time.Time `json:",omitempty"`
	Message  string    `json:",omitempty"`
	By       string    `json:",omitempty"`
	Manually bool
	AlertID  string `json:"alertId,omitempty"`
	AlertURL string `json:"alertUrl,omitempty"`
}

type PagedPolicy struct {
	Team  struct {
		Slug     string        `json:"slug,omitempty"`
		Name     string        `json:"name,omitempty"`
	}                              `json:"team,omitempty"`
	Policy  struct {
		Slug     string        `json:"slug,omitempty"`
		Name     string        `json:"name,omitempty"`
	}                              `json:"policy,omitempty"`
}


// Incident represents an incident on victorops
type Incident struct {
	AlertCount        int          `json:",omitempty"`
	CurrentPhase      string       `json:",omitempty"`
	EntityDisplayName string       `json:",omitempty"`
	EntityID          string       `json:",omitempty"`
	EntityState       string       `json:",omitempty"`
	EntityType        string       `json:",omitempty"`
	Host              string       `json:",omitempty"`
	IncidentNumber    string       `json:",omitempty"`
	LastAlertID       string       `json:",omitempty"`
	LastAlertTime     time.Time    `json:",omitempty"`
	Service           string       `json:",omitempty"`
	StartTime         time.Time    `json:",omitempty"`
	PagedTeams        []string     `json:",omitempty"`
	PagedUsers        []string     `json:",omitempty"`
	PagedPolicies     []PagedPolicy `json:",omitempty"`	
	Transitions       []Transition `json:",omitempty"`
}

// Incidents get a list of the currently open, acknowledged and
// recently resolved incidents
func (c *Client) Incidents() (incidents []Incident, err error) {
	resp, err := c.request(http.MethodGet, "v1/incidents", nil)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	var result = struct {
		Incidents []Incident
	}{}
	err = json.NewDecoder(resp.Body).Decode(&result)
	return result.Incidents, err
}

// Target User or EscalationPolicy
type Target struct {
	Type string `json:"type"`
	Slug string `json:"slug"`
}

type createIncidentBody struct {
	Summary  string   `json:"summary,omitempty"`
	Details  string   `json:"details,omitempty"`
	Username string   `json:"userName"`
	Targets  []Target `json:"targets"`
}

// CreatedIncident is an incident that was just created
type CreatedIncident struct {
	Number string `json:"incidentNumber,ommitempty"`
	Error  string `json:"error,ommitempty"`
}

// CreateIncident Create a new incident.
func (c *Client) CreateIncident(summary, details string, targets []Target) (incident CreatedIncident, err error) {
	var body = createIncidentBody{
		Username: c.User,
		Summary:  summary,
		Details:  details,
		Targets:  targets,
	}
	bts, err := json.Marshal(body)
	if err != nil {
		return
	}
	resp, err := c.request(http.MethodPost, "v1/incidents", bytes.NewBuffer(bts))
	if err != nil {
		return
	}
	defer resp.Body.Close()
	err = json.NewDecoder(resp.Body).Decode(&incident)
	return
}


type Reroute struct {
	IncidentNumber string                  `json:"incidentNumber,omitempty"`
	Targets        []Target                `json:"targets,omitempty"`
}
type rerouteIncidentBody struct {
	Username string                        `json:"userName,omitempty"`
	Reroutes []Reroute                     `json:"reroutes,omitempty"`
}

// RouteIncident reroute a existing incident.
func (c *Client) RerouteIncident(incidentNumber string, targets []Target) (responsejson string, err error) {
	var body = rerouteIncidentBody{
		Username: c.User,
		Reroutes: []Reroute{
			{IncidentNumber: incidentNumber, Targets: targets},
		},
	}
	bts, err := json.Marshal(body)
	if err != nil {
		return
	}
	resp, err := c.request(http.MethodPost, "v1/incidents/reroute", bytes.NewBuffer(bts))
	if err != nil {
		return
	}
	defer resp.Body.Close()
	// TODO: return it as a real json object - but for info at least return string - should contain statuses (success)
	responsebytes, _ := ioutil.ReadAll(resp.Body)
	responsejson = string(responsebytes)
	return
}


type incidentStateChangeRequest struct {
	Message   string   `json:"message,omitempty"`
	Incidents []string `json:"incidentNames,omitempty"`
	Username  string   `json:"userName"`
}

// IncidentStateResult represents a ACK or Resolve event on an incident
type IncidentStateResult struct {
	Number   string `json:"incidentNumber,ommitempty"`
	ID       string `json:"entityId,ommitempty"`
	Accepted bool   `json:"cmdAccepted"`
	Message  string `json:"message,ommitempty"`
}

func (c *Client) changeIncidents(incidentIDs []string, message, state string) (incidents []IncidentStateResult, err error) {
	var body = incidentStateChangeRequest{
		Message:   message,
		Username:  c.User,
		Incidents: incidentIDs,
	}
	bts, err := json.Marshal(body)
	if err != nil {
		return
	}
	resp, err := c.request(http.MethodPatch, "v1/incidents/"+state, bytes.NewBuffer(bts))
	if err != nil {
		return
	}
	defer resp.Body.Close()
	var result struct {
		Results []IncidentStateResult `json:"results"`
	}
	err = json.NewDecoder(resp.Body).Decode(&result)
	return result.Results, err
}

// Ack one or more incidents
func (c *Client) Ack(message string, incidentIDs ...string) ([]IncidentStateResult, error) {
	return c.changeIncidents(incidentIDs, message, "ack")
}

// Resolve one or more incidents
func (c *Client) Resolve(message string, incidentIDs ...string) ([]IncidentStateResult, error) {
	return c.changeIncidents(incidentIDs, message, "resolve")
}
