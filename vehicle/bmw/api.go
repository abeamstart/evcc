package bmw

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/evcc-io/evcc/util"
	"github.com/evcc-io/evcc/util/request"
	"golang.org/x/oauth2"
)

// https://github.com/bimmerconnected/bimmer_connected
// https://github.com/TA2k/ioBroker.bmw

const (
	ApiURI     = "https://b2vapi.bmwgroup.com/webapi/v1"
	CocoApiURI = "https://cocoapi.bmwgroup.com"
	XUserAgent = "android(v1.07_20200330);bmw;1.7.0(11152)"
)

// API is an api.Vehicle implementation for BMW cars
type API struct {
	*request.Helper
}

// NewAPI creates a new vehicle
func NewAPI(log *util.Logger, identity oauth2.TokenSource) *API {
	v := &API{
		Helper: request.NewHelper(log),
	}

	// replace client transport with authenticated transport
	v.Client.Transport = &oauth2.Transport{
		Source: identity,
		Base:   v.Client.Transport,
	}

	return v
}

// Vehicles implements returns the /user/vehicles api
func (v *API) Vehicles() ([]string, error) {
	var resp VehiclesResponse
	uri := fmt.Sprintf("%s/user/vehicles", ApiURI)
	// uri := fmt.Sprintf("%s/eadrax-vcs/v1/vehicles", CocoApiURI, vin)

	req, err := http.NewRequest(http.MethodGet, uri, nil)
	if err == nil {
		err = v.DoJSON(req, &resp)
	}

	var vehicles []string
	for _, v := range resp.Vehicles {
		vehicles = append(vehicles, v.VIN)
	}

	return vehicles, err
}

func init() {
	util.RedactHook = nil
}

// Status implements the /user/vehicles/<vin>/status api
func (v *API) Status(vin string) (VehicleStatus, error) {
	var resp VehiclesStatusResponse
	uri := fmt.Sprintf("%s/eadrax-vcs/v1/vehicles?apptimezone=60&appDateTime=%d", CocoApiURI, time.Now().Unix())

	req, err := request.New(http.MethodGet, uri, nil, map[string]string{
		"X-User-Agent": XUserAgent,
	})
	if err == nil {
		err = v.DoJSON(req, &resp)
	}

	v.Images(vin)

	if l := len(resp); l != 1 {
		return VehicleStatus{}, fmt.Errorf("unexpected length: %d", l)
	}

	return resp[0], err
}

// Images implements the /user/vehicles/<vin>/status api
func (v *API) Images(vin string) (VehicleStatus, error) {
	view := "VehicleStatus"
	uri := fmt.Sprintf("%s/eadrax-ics/v3/presentation/vehicles/%s/images?carView=%s", CocoApiURI, vin, view)

	req, err := request.New(http.MethodGet, uri, nil, map[string]string{
		"X-User-Agent": XUserAgent,
		"Accept":       "image/png",
	})

	var resp *http.Response
	if err == nil {
		resp, err = v.Do(req)
	}
	_ = resp
	os.Exit(1)

	return VehicleStatus{}, err
}
