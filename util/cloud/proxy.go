package cloud

type API int

const (
	_ API = iota
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
