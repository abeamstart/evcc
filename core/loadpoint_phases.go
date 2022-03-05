package core

import (
	"math"

	"github.com/evcc-io/evcc/api"
)

// TODO move measuredPhases to api/sync

// setMeasuredPhases provides synchronized access to measuredPhases
func (lp *LoadPoint) setMeasuredPhases(phases int) {
	lp.Lock()
	lp.measuredPhases = phases
	lp.Unlock()

	if phases == 0 {
		phases = lp.activePhases()
	}

	lp.publish("activePhases", phases)
}

// getMeasuredPhases provides synchronized access to measuredPhases
func (lp *LoadPoint) getMeasuredPhases() int {
	lp.Lock()
	defer lp.Unlock()
	return lp.measuredPhases
}

// assume 3p for switchable charger during startup
const unknownPhases = 3

func min(i ...int) int {
	v := math.MaxInt
	for _, i := range i {
		if i < v {
			v = i
		}
	}
	return v
}

func expect(phases int) int {
	if phases > 0 {
		return phases
	}
	return unknownPhases
}

// activePhases returns the number of expectedly active phases for the meter.
// If unknown for 1p3p chargers during startup it will assume 3p.
func (lp *LoadPoint) activePhases() int {
	physical := lp.GetPhases()
	vehicle := lp.getVehiclePhases()
	measured := lp.getMeasuredPhases()

	return min(expect(vehicle), expect(physical), expect(measured))
}

// maxActivePhases returns the maximum number of active phases for the meter.
func (lp *LoadPoint) maxActivePhases() int {
	physical := lp.GetPhases()
	measured := lp.getMeasuredPhases()
	vehicle := lp.getVehiclePhases()

	// during 1p or unknown config, 1p measured is not a restriction
	if physical <= 1 || vehicle == 1 {
		measured = 0
	}

	// if 1p3p supported then assume 3p
	if _, ok := lp.charger.(api.ChargePhases); ok {
		physical = 3
	}

	return min(expect(vehicle), expect(physical), expect(measured))
}

func (lp *LoadPoint) getVehiclePhases() int {
	if lp.vehicle != nil {
		return lp.vehicle.Phases()
	}

	return 0
}
