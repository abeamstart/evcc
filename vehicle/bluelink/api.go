package bluelink

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/evcc-io/evcc/util"
	"github.com/evcc-io/evcc/util/request"
	"github.com/evcc-io/evcc/util/transport"
)

const (
	resOK = "S" // auth fail: F
)

// ErrAuthFail indicates authorization failure
var ErrAuthFail = errors.New("authorization failed")

// API implements the Kia/Hyundai bluelink api.
// Based on https://github.com/Hacksore/bluelinky.
type API struct {
	*request.Helper
	baseURI  string
	identity Requester
}

type Requester interface {
	Request(*http.Request) error
	DeviceID() string
}

// New creates a new BlueLink API
func NewAPI(log *util.Logger, baseURI string, identity Requester, cache time.Duration) *API {
	v := &API{
		Helper:   request.NewHelper(log),
		baseURI:  strings.TrimSuffix(baseURI, "/api/v1/spa") + "/api",
		identity: identity,
	}

	// api is unbelievably slow when retrieving status
	v.Client.Timeout = 120 * time.Second

	v.Client.Transport = &transport.Decorator{
		Decorator: identity.Request,
		Base:      v.Client.Transport,
	}

	return v
}

type Vehicle struct {
	VIN, VehicleName, VehicleID string
}

func (v *API) Vehicles() ([]Vehicle, error) {
	var res VehiclesResponse

	uri := fmt.Sprintf("%s/v1/spa/vehicles", v.baseURI)
	err := v.GetJSON(uri, &res)

	return res.ResMsg.Vehicles, err
}

// StatusLatest retrieves the latest server-side status
func (v *API) StatusLatest(vid string) (StatusLatestResponse, error) {
	var res StatusLatestResponse

	uri := fmt.Sprintf("%s/v1/spa/vehicles/%s/status/latest", v.baseURI, vid)
	err := v.GetJSON(uri, &res)
	if err == nil && res.RetCode != resOK {
		err = fmt.Errorf("unexpected response: %s", res.RetCode)
	}

	return res, err
}

// StatusPartial refreshes the status
func (v *API) StatusPartial(vid string) (StatusResponse, error) {
	var res StatusResponse

	uri := fmt.Sprintf("%s/v1/spa/vehicles/%s/status", v.baseURI, vid)
	err := v.GetJSON(uri, &res)
	if err == nil && res.RetCode != resOK {
		err = fmt.Errorf("unexpected response: %s", res.RetCode)
	}

	return res, err
}

const (
	ActionCharge      = "charge"
	ActionChargeStart = "start"
	ActionChargeStop  = "stop"
)

// Action implements vehicle actions
// TODO add pin
func (v *API) Action(vid, action, value string) error {
	uri := fmt.Sprintf("%s/v2/spa/vehicles/%s/control/%s", v.baseURI, vid, action)

	body := struct {
		Action   string `json:"action"`
		DeviceId string `json:"deviceId"`
	}{
		Action:   value,
		DeviceId: v.identity.DeviceID(),
	}

	req, err := request.New(http.MethodPost, uri, request.MarshalJSON(body), request.JSONEncoding)

	if err == nil {
		var resp *http.Response
		if resp, err = v.Do(req); err == nil {
			resp.Body.Close()
		}
	}

	return err
}
