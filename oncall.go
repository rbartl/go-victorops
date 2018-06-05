package victorops

import (
	"encoding/json"
	"net/http"
	"time"
	"io/ioutil"
	"log"
	"fmt"
	"bytes"
	"strings"
	"strconv"
)

type UnixTimestamp struct {
 time.Time
}

func (t *UnixTimestamp) UnmarshalJSON(b []byte) error {
    i64, _ := strconv.ParseInt(string(b), 10, 64)
    t.Time=time.Unix(i64/1000,0)
    return nil
}

func (t *UnixTimestamp) MarshalJSON() ([]byte, error) {
        i64 := t.UnixNano()
        str := strconv.FormatInt(i64,10)
	buffer := bytes.NewBufferString(str)
	return buffer.Bytes(), nil
}



// Incident represents an incident on victorops
type TeamScheduleBlock struct {
	Team  struct {
		Name          string       `json:",omitempty"`
		Slug          string       `json:",omitempty"`
	}                              `json:",omitempty"`
	Schedule          []Schedule   `json:",omitempty"`
	Overrides         []Override   `json:",omitempty"`
}

type Schedule struct {
	OnCallUser  struct {
		Username      string       `json:",omitempty"`
	}                              `json:",omitempty"`
	OverrideOnCallUser  struct {
		Username      string       `json:",omitempty"`
	}                              `json:",omitempty"`
	RotationName      string       `json:",omitempty"`
	OnCallType        string       `json:",omitempty"`
	ShiftName         string       `json:",omitempty"`
	ShiftRoll         string       `json:",omitempty"`
	Rolls			  []Roll  	   `json:",omitempty"`
}

type Override struct {
	OrigOnCallUser  struct {
		Username      string       `json:",omitempty"`
	}                              `json:",omitempty"`
	OverrideOnCallUser  struct {
		Username      string       `json:",omitempty"`
	}                              `json:",omitempty"`
	Start             time.Time    `json:",omitempty"`
	End               time.Time    `json:",omitempty"`
}

type Roll struct {
	Start             time.Time    `json:",omitempty"`
	Isroll            bool         `json:",omitempty"`
	End               time.Time    `json:",omitempty"`
	OnCallUser  struct {
		Username      string       `json:",omitempty"`
	}                              `json:",omitempty"`
}

// Incidents get a list of the currently open, acknowledged and
// recently resolved incidents
func (c *Client) TeamSchedules() (schedules []TeamScheduleBlock, err error) {
	resp, err := c.request(http.MethodGet, "v2/team/ops/oncall/schedule", nil)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	var result = struct {
		Schedules []TeamScheduleBlock
	}{}
	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(string(bodyBytes))
	err = json.Unmarshal(bodyBytes, &result)
	if err != nil {
		log.Fatal(err)
	}
	return result.Schedules, err
}


type createOverrideRequest struct {
	Start       string     `json:"start,omitempty"`
	End         string     `json:"end,omitempty"`
	Timezone    string     `json:"timezone,omitempty"`
	Username    string     `json:"username,omitempty"`
}

func (c *Client) CreateOverride(start, end time.Time, username string) (err error) {


	var body = createOverrideRequest{
		Start:  start.Format("2006-01-02T15:04:05.000+0000"),
		End: 	end.Format("2006-01-02T15:04:05.000+0000"),
		Timezone : start.Location().String(),
		Username : username,
	}
	bts, err := json.Marshal(body)
	if err != nil {
		return
	}
	resp, err := c.privaterequest(http.MethodPost, "v2/org/netconomy/overrides", bytes.NewBuffer(bts))

	if err != nil {
		if strings.Contains(err.Error(), "conflicting") {
			// if the error contains conflicting ovrerride schedules they have already been created and thats ok
			return nil
		}
		log.Fatal(err)
		return
	}
	defer resp.Body.Close()
	var result = struct {
		Schedules []TeamScheduleBlock
	}{}

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	err = json.Unmarshal(bodyBytes, &result)
	if err != nil {
		log.Fatal(err)
	}
	return nil
}

// https://portal.victorops.com/api/v2/org/netconomy/overrides
func (c *Client) ListOverrides() (Overrides PrivOverrides, err error) {


	resp, err := c.privaterequest(http.MethodGet, "v2/org/netconomy/overrides", nil)

	if err != nil {
		log.Fatal(err)
		return
	}
	defer resp.Body.Close()
/*	var result = struct {
		Overrides []TeamScheduleBlock
	}{}
*/
	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	err = json.Unmarshal(bodyBytes, &Overrides)
	if err != nil {
		log.Fatal(err)
	}
	return Overrides, nil
}


type PrivOverrides struct {
	Overrides   []PrivOverride  `json:"overrides,omitempty"`
}


type PrivOverride struct {
	Id          int      `json:"id,omitempty"`
	User  struct {
		Username      string       `json:"username,omitempty"`
		FirstName     string       `json:"firstName,omitempty"`
		LastName      string       `json:"lastName,omitempty"`
	}                              `json:"user,omitempty"`
	Start       UnixTimestamp   `json:"start,omitempty"`
	End         UnixTimestamp   `json:"end,omitempty"`
	Timezone    string   `json:"timezone,omitempty"`
	Assignments    []PrivAssignment   `json:"assignments,omitempty"`
}
type PrivAssignment struct {
	Id          int      `json:"id,omitempty"`
	User  struct {
		Username      string       `json:"username,omitempty"`
		FirstName     string       `json:"firstName,omitempty"`
		LastName      string       `json:"lastName,omitempty"`
	}                              `json:"user,omitempty"`
	Team  struct {
		Slug      string            `json:"slug,omitempty"`
		Name     string            `json:"name,omitempty"`
	}                              `json:"team,omitempty"`
	Policy  struct {
		Slug      string            `json:"slug,omitempty"`
		Name     string            `json:"name,omitempty"`
	}                              `json:"policy,omitempty"`
}



type configureOverrideRequest struct {
	PolicySlug  string     `json:"policySlug,omitempty"`
	Username    string     `json:"username,omitempty"`
}


func (c *Client) ConfigureOverride(overrideid int, policyslug string, username string) (err error) {

	var body = configureOverrideRequest{
		PolicySlug:   policyslug,
		Username : username,
	}
	bts, err := json.Marshal(body)
	if err != nil {
		return
	}
	resp, err := c.privaterequest(http.MethodPut, "v2/org/netconomy/overrides/" + strconv.Itoa(overrideid), bytes.NewBuffer(bts))

	if err != nil {
		log.Fatal(err)
		return
	}
	defer resp.Body.Close()
	var result = struct {
		Schedules []TeamScheduleBlock
	}{}

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	err = json.Unmarshal(bodyBytes, &result)
	if err != nil {
		log.Fatal(err)
	}
	return nil
}


