package cloud

import (
	"encoding/gob"
	"time"

	"github.com/evcc-io/evcc/api"
	"github.com/evcc-io/evcc/core/loadpoint"
)

type ApiCall int

const (
	_ ApiCall = iota

	// site
	Healthy
	SetPrioritySoC

	// loadpoint
	Name
	HasChargeMeter
	GetStatus
	GetMode
	SetMode
	GetTargetSoC
	SetTargetSoC
	GetMinSoC
	SetMinSoC
	GetPhases
	SetPhases
	SetTargetCharge
	GetChargePower
	GetMinCurrent
	SetMinCurrent
	GetMaxCurrent
	SetMaxCurrent
	GetMinPower
	GetMaxPower
	GetRemainingDuration
	GetRemainingEnergy
	RemoteControl
)

func init() {
	RegisterTypes()
}

func RegisterTypes() {
	gob.Register(api.ModeEmpty)
	gob.Register(api.StatusNone)
	gob.Register(loadpoint.RemoteEnable)
	gob.Register(time.Duration(0))
	gob.Register(time.Time{})
}
