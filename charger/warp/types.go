package warp

import "time"

const (
	RootTopic = "warp"
	Timeout   = 30 * time.Second
)

// https://www.warp-charger.com/api.html#evse_state
type EvseState struct {
	Iec61851State          int   `json:"iec61851_state"`
	VehicleState           int   `json:"vehicle_state"`
	ChargeRelease          int   `json:"charge_release"`
	ContactorState         int   `json:"contactor_state"`
	ContactorError         int   `json:"contactor_error"`
	AllowedChargingCurrent int64 `json:"allowed_charging_current"`
	ErrorState             int   `json:"error_state"`
	LockState              int   `json:"lock_state"`
	TimeSinceStateChange   int64 `json:"time_since_state_change"`
	Uptime                 int64 `json:"uptime"`
}

// https://www.warp-charger.com/api.html#evse_low_level_state
type LowLevelState struct {
	LedState       int   `json:"led_state"`
	CpPwmDutyCycle int   `json:"cp_pwm_duty_cycle"`
	AdcValues      []int `json:"adc_values"`
	Voltages       []int
	Resistances    []int
	Gpio           []bool
}

// https://www.warp-charger.com/api.html#meter_state
type MeterState struct {
	State           int     `json:"state"` // Warp 1 only
	Power           float64 `json:"power"`
	EnergyRel       float64 `json:"energy_rel"`
	EnergyAbs       float64 `json:"energy_abs"`
	PhasesActive    []bool  `json:"phases_active"`
	PhasesConnected []bool  `json:"phases_connected"`
}

type NfcTag struct {
	Type     int    `json:"tag_type"`
	ID       []byte `json:"tag_id"`
	LastSeen int64  `json:"last_seen"`
}
