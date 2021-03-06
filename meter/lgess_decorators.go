package meter

// Code generated by github.com/evcc-io/evcc/cmd/tools/decorate.go. DO NOT EDIT.

import (
	"github.com/evcc-io/evcc/api"
)

func decorateLgEss(base *LgEss, meterEnergy func() (float64, error), battery func() (float64, error)) api.Meter {
	switch {
	case battery == nil && meterEnergy == nil:
		return base

	case battery == nil && meterEnergy != nil:
		return &struct {
			*LgEss
			api.MeterEnergy
		}{
			LgEss: base,
			MeterEnergy: &decorateLgEssMeterEnergyImpl{
				meterEnergy: meterEnergy,
			},
		}

	case battery != nil && meterEnergy == nil:
		return &struct {
			*LgEss
			api.Battery
		}{
			LgEss: base,
			Battery: &decorateLgEssBatteryImpl{
				battery: battery,
			},
		}

	case battery != nil && meterEnergy != nil:
		return &struct {
			*LgEss
			api.Battery
			api.MeterEnergy
		}{
			LgEss: base,
			Battery: &decorateLgEssBatteryImpl{
				battery: battery,
			},
			MeterEnergy: &decorateLgEssMeterEnergyImpl{
				meterEnergy: meterEnergy,
			},
		}
	}

	return nil
}

type decorateLgEssBatteryImpl struct {
	battery func() (float64, error)
}

func (impl *decorateLgEssBatteryImpl) SoC() (float64, error) {
	return impl.battery()
}

type decorateLgEssMeterEnergyImpl struct {
	meterEnergy func() (float64, error)
}

func (impl *decorateLgEssMeterEnergyImpl) TotalEnergy() (float64, error) {
	return impl.meterEnergy()
}
