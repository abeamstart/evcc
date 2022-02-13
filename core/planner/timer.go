package planner

import (
	"math"
	"time"

	"github.com/evcc-io/evcc/api"
	"github.com/evcc-io/evcc/core/soc"
	"github.com/evcc-io/evcc/util"
)

const (
	deviation = 30 * time.Minute
)

// Timer is the target charging handler
type Timer struct {
	Adapter
	log       *util.Logger
	current   float64
	finishAt  time.Time
	active    bool
	validated bool
}

// NewTimer creates a time planner
func NewTimer(log *util.Logger, api Adapter) *Timer {
	lp := &Timer{
		log:     log,
		Adapter: api,
	}

	return lp
}

// MustValidateDemand resets the flag for detecting if DemandActive has been called
func (p *Timer) MustValidateDemand() {
	if p == nil {
		return
	}

	p.validated = false
}

// DemandValidated returns if DemandActive has been called
func (p *Timer) DemandValidated() bool {
	if p == nil {
		return false
	}

	return p.validated
}

func (p *Timer) RemainingDuration() time.Duration {
	se := p.SocEstimator()
	if se == nil {
		return 0
	}

	power := p.GetMaxPower()
	if p.active {
		power *= p.current / p.GetMaxCurrent()
	}

	return se.RemainingChargeDuration(power, p.GetTargetSoC())
}

// Stop stops the target charging request
func (p *Timer) Stop() {
	if p == nil {
		return
	}

	if p.active {
		p.active = false
		p.Publish("targetTimeActive", p.active)
		p.log.DEBUG.Println("target charging: disable")
	}
}

// // Reset resets the target charging request
// func (p *Timer) Reset() {
// 	if p == nil {
// 		return
// 	}

// 	p.Set(time.Time{})
// 	p.Stop()
// }

// Active calculates remaining charge duration and returns true if charge start is required to achieve target soc in time
func (p *Timer) Active() bool {
	if p == nil || p.GetTargetTime().IsZero() {
		return false
	}

	// demand validation has been called
	p.validated = true

	// power
	power := p.GetMaxPower()
	if p.active {
		power *= p.current / p.GetMaxCurrent()
	}

	se := p.SocEstimator()
	if se == nil {
		return false
	}

	// time
	remainingDuration := time.Duration(float64(se.AssumedChargeDuration(p.GetTargetSoC(), power)) / soc.ChargeEfficiency)
	p.finishAt = time.Now().Add(remainingDuration).Round(time.Minute)

	p.log.DEBUG.Printf("estimated charge duration: %v to %d%% at %.0fW", remainingDuration.Round(time.Minute), p.GetTargetSoC(), power)
	if p.active {
		p.log.DEBUG.Printf("projected end: %v", p.finishAt)
		p.log.DEBUG.Printf("desired finish time: %v", p.GetTargetTime())
	} else {
		p.log.DEBUG.Printf("projected start: %v", p.GetTargetTime().Add(-remainingDuration))
	}

	// timer charging is already active- only deactivate once charging has stopped
	if p.active {
		if time.Now().After(p.GetTargetTime()) && p.GetStatus() != api.StatusC {
			p.Stop()
		}

		return p.active
	}

	// check if charging need be activated
	if active := p.finishAt.After(p.GetTargetTime()); active {
		p.active = active
		p.Publish("targetTimeActive", p.active)

		p.current = p.GetMaxCurrent()
		p.log.INFO.Printf("target charging active for %v: projected %v (%v remaining)", p.GetTargetTime(), p.finishAt, remainingDuration.Round(time.Minute))
	}

	return p.active
}

// Handle adjusts current up/down to achieve desired target time taking.
func (p *Timer) Handle() float64 {
	action := "steady"

	switch {
	case p.finishAt.Before(p.GetTargetTime().Add(-deviation)):
		p.current--
		action = "slowdown"

	case p.finishAt.After(p.GetTargetTime()):
		p.current++
		action = "speedup"
	}

	p.current = math.Max(math.Min(p.current, p.GetMaxCurrent()), p.GetMinCurrent())
	p.log.DEBUG.Printf("target charging: %s (%.3gA)", action, p.current)

	return p.current
}
