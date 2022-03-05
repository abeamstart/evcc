package core

import (
	"strings"
	"testing"
	"time"

	evbus "github.com/asaskevich/EventBus"
	"github.com/benbjohnson/clock"
	"github.com/evcc-io/evcc/api"
	"github.com/evcc-io/evcc/mock"
	"github.com/evcc-io/evcc/util"
	"github.com/golang/mock/gomock"
)

type testCase struct {
	// capable=0 signals 1p3p as set during loadpoint init
	// physical/vehicle=0 signals unknown
	// measuredPhases<>0 signals previous measurement
	capable, physical, vehicle, measuredPhases, actExpected, maxExpected int
	// scaling expectation: d=down, u=up, du=both
	scale string
}

var (
	phaseTests = []testCase{
		// 1p
		{1, 1, 0, 0, 1, 1, ""},
		{1, 1, 0, 1, 1, 1, ""},
		{1, 1, 1, 0, 1, 1, ""},
		{1, 1, 2, 0, 1, 1, ""},
		{1, 1, 3, 0, 1, 1, ""},
		// 3p
		{3, 3, 0, 0, unknownPhases, 3, ""},
		{3, 3, 0, 1, 1, 1, ""},
		{3, 3, 0, 2, 2, 2, ""},
		{3, 3, 0, 3, 3, 3, ""},
		{3, 3, 1, 0, 1, 1, ""},
		{3, 3, 2, 0, 2, 2, ""},
		{3, 3, 3, 0, 3, 3, ""},
		// 1p3p initial
		{0, 0, 0, 0, unknownPhases, 3, "du"},
		{0, 0, 0, 1, 1, 3, "u"},
		{0, 0, 0, 2, 2, 3, "du"},
		{0, 0, 0, 3, 3, 3, "du"},
		{0, 0, 1, 0, 1, 1, ""},
		{0, 0, 2, 0, 2, 2, "du"},
		{0, 0, 3, 0, 3, 3, "du"},
		// 1p3p, 1 currently active
		{0, 1, 0, 0, 1, 3, "u"},
		{0, 1, 0, 1, 1, 3, "u"},
		// {0, 1, 0, 2, 2,2,"u"}, // 2p active > 1p configured must not happen
		// {0, 1, 0, 3, 3,3,"u"}, // 3p active > 1p configured must not happen
		{0, 1, 1, 0, 1, 1, ""},
		{0, 1, 2, 0, 1, 2, "u"},
		{0, 1, 3, 0, 1, 3, "u"},
		// 1p3p, 3 currently active
		{0, 3, 0, 0, unknownPhases, 3, "d"},
		{0, 3, 0, 1, 1, 1, ""},
		{0, 3, 0, 2, 2, 2, "d"},
		{0, 3, 0, 3, 3, 3, "d"},
		{0, 3, 1, 0, 1, 1, ""},
		{0, 3, 2, 0, 2, 2, "d"},
		{0, 3, 3, 0, 3, 3, "d"},
	}
)

func testScale(t *testing.T, lp *LoadPoint, power float64, direction string, tc testCase) {
	act := lp.activePhases()
	max := lp.maxActivePhases()
	scaled := lp.pvScalePhases(power, minA, maxA)

	if strings.Contains(tc.scale, direction[0:1]) {
		if !scaled {
			t.Errorf("%v act=%d max=%d missing scale %s", tc, act, max, direction)
		}
	} else if scaled {
		t.Errorf("%v act=%d max=%d unexpected scale %s", tc, act, max, direction)
	}
}

func TestPvScalePhases(t *testing.T) {
	clock := clock.NewMock()
	ctrl := gomock.NewController(t)

	for _, tc := range phaseTests {
		t.Log(tc)

		plainCharger := mock.NewMockCharger(ctrl)
		plainCharger.EXPECT().Enabled().Return(true, nil)
		plainCharger.EXPECT().MaxCurrent(int64(minA)).Return(nil) // MaxCurrentEx not implemented

		// 1p3p
		var phaseCharger *mock.MockChargePhases
		if tc.capable == 0 {
			phaseCharger = mock.NewMockChargePhases(ctrl)
		}

		vehicle := mock.NewMockVehicle(ctrl)
		vehicle.EXPECT().Phases().Return(tc.vehicle).MinTimes(1)

		lp := &LoadPoint{
			log:         util.NewLogger("foo"),
			bus:         evbus.New(),
			clock:       clock,
			chargeMeter: &Null{},            // silence nil panics
			chargeRater: &Null{},            // silence nil panics
			chargeTimer: &Null{},            // silence nil panics
			progress:    NewProgress(0, 10), // silence nil panics
			wakeUpTimer: NewTimer(),         // silence nil panics
			Mode:        api.ModeNow,
			MinCurrent:  minA,
			MaxCurrent:  maxA,
			vehicle:     vehicle,
			Phases:      tc.physical,
		}

		if phaseCharger != nil {
			lp.charger = struct {
				*mock.MockCharger
				*mock.MockChargePhases
			}{
				plainCharger, phaseCharger,
			}
		} else {
			lp.charger = struct {
				*mock.MockCharger
			}{
				plainCharger,
			}
		}

		attachListeners(t, lp)

		lp.setMeasuredPhases(tc.measuredPhases)
		if tc.measuredPhases > 0 && tc.vehicle > 0 {
			t.Fatalf("%v invalid test case", tc)
		}

		if lp.GetPhases() != tc.physical {
			t.Error("wrong phases", lp.GetPhases(), tc.physical)
		}

		if phs := lp.activePhases(); phs != tc.actExpected {
			t.Errorf("expected active %d, got %d", tc.actExpected, phs)
		}
		if phs := lp.maxActivePhases(); phs != tc.maxExpected {
			t.Errorf("expected max %d, got %d", tc.maxExpected, phs)
		}
		ctrl.Finish()

		// scaling
		if phaseCharger != nil {
			// scale down
			min1p := 1 * minA * Voltage
			lp.phaseTimer = time.Time{}

			plainCharger.EXPECT().Enable(false).Return(nil).MaxTimes(1)
			phaseCharger.EXPECT().Phases1p3p(1).Return(nil).MaxTimes(1)

			testScale(t, lp, min1p, "down", tc)
			ctrl.Finish()

			// scale up
			min3p := float64(tc.maxExpected) * minA * Voltage
			lp.phaseTimer = time.Time{}

			// reset to initial state
			_ = lp.SetPhases(tc.physical)
			lp.setMeasuredPhases(tc.measuredPhases)

			plainCharger.EXPECT().Enable(false).Return(nil).MaxTimes(1)
			phaseCharger.EXPECT().Phases1p3p(3).Return(nil).MaxTimes(1)

			testScale(t, lp, min3p, "up", tc)
			ctrl.Finish()
		}
	}
}

func TestPvScalePhasesTimer(t *testing.T) {
	ctrl := gomock.NewController(t)
	charger := &struct {
		*mock.MockCharger
		*mock.MockChargePhases
	}{
		mock.NewMockCharger(ctrl),
		mock.NewMockChargePhases(ctrl),
	}

	dt := time.Minute
	Voltage = 230 // V

	tc := []struct {
		desc                   string
		phases, measuredPhases int
		availablePower         float64
		toPhases               int
		res                    bool
		prepare                func(lp *LoadPoint)
	}{
		// switch up from 1p/1p configured/active
		{"1/1->3, not enough power", 1, 1, 0, 1, false, nil},
		{"1/1->3, kickoff", 1, 1, 3 * Voltage * minA, 1, false, func(lp *LoadPoint) {
			lp.phaseTimer = time.Time{}
		}},
		{"1/1->3, timer running", 1, 1, 3 * Voltage * minA, 1, false, func(lp *LoadPoint) {
			lp.phaseTimer = lp.clock.Now()
		}},
		{"1/1->3, timer elapsed", 1, 1, 3 * Voltage * minA, 3, true, func(lp *LoadPoint) {
			lp.phaseTimer = lp.clock.Now().Add(-dt)
		}},

		// omit to switch up (again) from 3p/1p configured/active
		{"3/1->3, not enough power", 3, 1, 0, 3, false, nil},
		{"3/1->3, kickoff", 3, 1, 3 * Voltage * minA, 3, false, func(lp *LoadPoint) {
			lp.phaseTimer = time.Time{}
		}},
		{"3/1->3, timer running", 3, 1, 3 * Voltage * minA, 3, false, func(lp *LoadPoint) {
			lp.phaseTimer = lp.clock.Now()
		}},
		{"3/1->3, timer elapsed", 3, 1, 3 * Voltage * minA, 3, false, func(lp *LoadPoint) {
			lp.phaseTimer = lp.clock.Now().Add(-dt)
		}},

		// omit to switch down from 3p/1p configured/active
		{"3/1->1, not enough power", 3, 1, 0, 3, false, nil},
		{"3/1->1, kickoff", 3, 1, 1 * Voltage * minA, 3, false, func(lp *LoadPoint) {
			lp.phaseTimer = time.Time{}
		}},
		{"3/1->1, timer running", 3, 1, 1 * Voltage * minA, 3, false, func(lp *LoadPoint) {
			lp.phaseTimer = lp.clock.Now()
		}},
		{"3/1->1, timer elapsed", 3, 1, 1 * Voltage * minA, 3, false, func(lp *LoadPoint) {
			lp.phaseTimer = lp.clock.Now().Add(-dt)
		}},

		// switch down from 3p/3p configured/active
		{"3/3->1, enough power", 3, 3, 1 * Voltage * maxA, 3, false, nil},
		{"3/3->1, kickoff", 3, 3, 1 * Voltage * maxA, 3, false, func(lp *LoadPoint) {
			lp.phaseTimer = time.Time{}
		}},
		{"3/3->1, timer running", 3, 3, 1 * Voltage * maxA, 3, false, func(lp *LoadPoint) {
			lp.phaseTimer = lp.clock.Now()
		}},
		{"3/3->1, timer elapsed", 3, 3, 1 * Voltage * maxA, 1, true, func(lp *LoadPoint) {
			lp.phaseTimer = lp.clock.Now().Add(-dt)
		}},

		// error states from 1p/3p misconfig - no correction for time being (stay at 1p)
		{"1/3->1, enough power", 1, 3, 1 * Voltage * maxA, 1, false, nil},
		{"1/3->1, kickoff, correct phase setting", 1, 3, 1 * Voltage * maxA, 1, false, func(lp *LoadPoint) {
			lp.phaseTimer = time.Time{}
		}},
		{"1/3->1, timer running, correct phase setting", 1, 3, 1 * Voltage * maxA, 1, false, func(lp *LoadPoint) {
			lp.phaseTimer = lp.clock.Now()
		}},
		{"1/3->1, switch not executed", 1, 3, 1 * Voltage * maxA, 1, false, func(lp *LoadPoint) {
			lp.phaseTimer = lp.clock.Now().Add(-dt)
		}},
	}

	for _, tc := range tc {
		t.Logf("%+v", tc)
		clock := clock.NewMock()
		clock.Add(time.Hour) // avoid time.IsZero

		lp := &LoadPoint{
			log:            util.NewLogger("foo"),
			clock:          clock,
			charger:        charger,
			MinCurrent:     minA,
			MaxCurrent:     maxA,
			Phases:         tc.phases,
			measuredPhases: tc.measuredPhases,
			Enable: ThresholdConfig{
				Delay: dt,
			},
			Disable: ThresholdConfig{
				Delay: dt,
			},
		}

		if tc.prepare != nil {
			tc.prepare(lp)
		}

		if tc.res {
			charger.MockChargePhases.EXPECT().Phases1p3p(tc.toPhases).Return(nil)
		}

		res := lp.pvScalePhases(tc.availablePower, minA, maxA)

		switch {
		case tc.res != res:
			t.Errorf("expected %v, got %v", tc.res, res)
		case lp.GetPhases() != tc.toPhases:
			t.Errorf("expected %dp, got %dp", tc.toPhases, lp.GetPhases())
		}
	}
}
