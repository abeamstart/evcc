package charger

// Code generated by github.com/evcc-io/evcc/cmd/tools/decorate.go. DO NOT EDIT.

import (
	"github.com/evcc-io/evcc/api"
)

func decoratePhoenixEMEth(base *PhoenixEMEth, meter func() (float64, error), meterEnergy func() (float64, error), meterCurrent func() (float64, float64, float64, error)) api.Charger {
	switch {
	case meter == nil && meterCurrent == nil && meterEnergy == nil:
		return base

	case meter != nil && meterCurrent == nil && meterEnergy == nil:
		return &struct {
			*PhoenixEMEth
			api.Meter
		}{
			PhoenixEMEth: base,
			Meter: &decoratePhoenixEMEthMeterImpl{
				meter: meter,
			},
		}

	case meter == nil && meterCurrent == nil && meterEnergy != nil:
		return &struct {
			*PhoenixEMEth
			api.MeterEnergy
		}{
			PhoenixEMEth: base,
			MeterEnergy: &decoratePhoenixEMEthMeterEnergyImpl{
				meterEnergy: meterEnergy,
			},
		}

	case meter != nil && meterCurrent == nil && meterEnergy != nil:
		return &struct {
			*PhoenixEMEth
			api.Meter
			api.MeterEnergy
		}{
			PhoenixEMEth: base,
			Meter: &decoratePhoenixEMEthMeterImpl{
				meter: meter,
			},
			MeterEnergy: &decoratePhoenixEMEthMeterEnergyImpl{
				meterEnergy: meterEnergy,
			},
		}

	case meter == nil && meterCurrent != nil && meterEnergy == nil:
		return &struct {
			*PhoenixEMEth
			api.MeterCurrent
		}{
			PhoenixEMEth: base,
			MeterCurrent: &decoratePhoenixEMEthMeterCurrentImpl{
				meterCurrent: meterCurrent,
			},
		}

	case meter != nil && meterCurrent != nil && meterEnergy == nil:
		return &struct {
			*PhoenixEMEth
			api.Meter
			api.MeterCurrent
		}{
			PhoenixEMEth: base,
			Meter: &decoratePhoenixEMEthMeterImpl{
				meter: meter,
			},
			MeterCurrent: &decoratePhoenixEMEthMeterCurrentImpl{
				meterCurrent: meterCurrent,
			},
		}

	case meter == nil && meterCurrent != nil && meterEnergy != nil:
		return &struct {
			*PhoenixEMEth
			api.MeterCurrent
			api.MeterEnergy
		}{
			PhoenixEMEth: base,
			MeterCurrent: &decoratePhoenixEMEthMeterCurrentImpl{
				meterCurrent: meterCurrent,
			},
			MeterEnergy: &decoratePhoenixEMEthMeterEnergyImpl{
				meterEnergy: meterEnergy,
			},
		}

	case meter != nil && meterCurrent != nil && meterEnergy != nil:
		return &struct {
			*PhoenixEMEth
			api.Meter
			api.MeterCurrent
			api.MeterEnergy
		}{
			PhoenixEMEth: base,
			Meter: &decoratePhoenixEMEthMeterImpl{
				meter: meter,
			},
			MeterCurrent: &decoratePhoenixEMEthMeterCurrentImpl{
				meterCurrent: meterCurrent,
			},
			MeterEnergy: &decoratePhoenixEMEthMeterEnergyImpl{
				meterEnergy: meterEnergy,
			},
		}
	}

	return nil
}

type decoratePhoenixEMEthMeterImpl struct {
	meter func() (float64, error)
}

func (impl *decoratePhoenixEMEthMeterImpl) CurrentPower() (float64, error) {
	return impl.meter()
}

type decoratePhoenixEMEthMeterCurrentImpl struct {
	meterCurrent func() (float64, float64, float64, error)
}

func (impl *decoratePhoenixEMEthMeterCurrentImpl) Currents() (float64, float64, float64, error) {
	return impl.meterCurrent()
}

type decoratePhoenixEMEthMeterEnergyImpl struct {
	meterEnergy func() (float64, error)
}

func (impl *decoratePhoenixEMEthMeterEnergyImpl) TotalEnergy() (float64, error) {
	return impl.meterEnergy()
}
