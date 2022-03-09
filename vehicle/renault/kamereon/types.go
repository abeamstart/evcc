package kamereon

type Response struct {
	Accounts     []Account `json:"accounts"`     // /commerce/v1/persons/%s
	AccessToken  string    `json:"accessToken"`  // /commerce/v1/accounts/%s/kamereon/token
	VehicleLinks []Vehicle `json:"vehicleLinks"` // /commerce/v1/accounts/%s/vehicles
	Data         Data      `json:"data"`         // /commerce/v1/accounts/%s/kamereon/kca/car-adapter/vX/cars/%s/...
}

type Account struct {
	AccountID string `json:"accountId"`
}

type Vehicle struct {
	Brand  string `json:"brand"`
	VIN    string `json:"vin"`
	Status string `json:"status"`
}

type Data struct {
	Attributes attributes `json:"attributes"`
}

type attributes struct {
	// battery-status
	Timestamp          string  `json:"timestamp"`
	ChargingStatus     float32 `json:"chargingStatus"`
	InstantaneousPower int     `json:"instantaneousPower"`
	RangeHvacOff       int     `json:"rangeHvacOff"`
	BatteryAutonomy    int     `json:"batteryAutonomy"`
	BatteryLevel       int     `json:"batteryLevel"`
	BatteryTemperature int     `json:"batteryTemperature"`
	PlugStatus         int     `json:"plugStatus"`
	LastUpdateTime     string  `json:"lastUpdateTime"`
	ChargePower        int     `json:"chargePower"`
	RemainingTime      *int    `json:"chargingRemainingTime"`
	// hvac-status
	ExternalTemperature float64 `json:"externalTemperature"`
	HvacStatus          string  `json:"hvacStatus"`
	// cockpit
	TotalMileage float64 `json:"totalMileage"`
}
