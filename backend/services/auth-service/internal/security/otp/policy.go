package otp

import (
	"errors"
	"time"
)

type Policy struct {
	Length             int
	TTL                time.Duration
	MaxAttempts        int
	ResendCooldown     time.Duration
	SendLimitWindow    time.Duration
	SendLimitPerWindow int
	DailyLimit         int
	VerifyIPWindow     time.Duration
	VerifyIPLimit      int
}

func DefaultPolicy() Policy {
	return Policy{
		Length:             6,
		TTL:                5 * time.Minute,
		MaxAttempts:        5,
		ResendCooldown:     time.Minute,
		SendLimitWindow:    15 * time.Minute,
		SendLimitPerWindow: 3,
		DailyLimit:         10,
		VerifyIPWindow:     5 * time.Minute,
		VerifyIPLimit:      20,
	}
}

func (p Policy) Validate() error {
	if p.Length < 6 || p.Length > 10 {
		return errors.New("otp length must be between 6 and 10")
	}
	if p.TTL <= 0 {
		return errors.New("otp ttl must be greater than zero")
	}
	if p.MaxAttempts <= 0 {
		return errors.New("otp max attempts must be greater than zero")
	}
	if p.ResendCooldown <= 0 {
		return errors.New("otp resend cooldown must be greater than zero")
	}
	if p.SendLimitWindow <= 0 {
		return errors.New("otp send limit window must be greater than zero")
	}
	if p.SendLimitPerWindow <= 0 {
		return errors.New("otp send limit per window must be greater than zero")
	}
	if p.DailyLimit <= 0 {
		return errors.New("otp daily limit must be greater than zero")
	}
	if p.VerifyIPWindow <= 0 {
		return errors.New("otp verify ip window must be greater than zero")
	}
	if p.VerifyIPLimit <= 0 {
		return errors.New("otp verify ip limit must be greater than zero")
	}
	return nil
}
