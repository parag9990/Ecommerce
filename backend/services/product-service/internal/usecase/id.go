package usecase

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"
)

type IDGenerator interface {
	NewProductID() string
	NewVariantID() string
	NewImageID() string
}

type ProductEventIDGenerator interface {
	NewProductEventID() string
}

type Clock interface {
	Now() time.Time
}

type RandomIDGenerator struct{}

func NewRandomIDGenerator() RandomIDGenerator {
	return RandomIDGenerator{}
}

func (RandomIDGenerator) NewProductID() string {
	return randomID("prod")
}

func (RandomIDGenerator) NewVariantID() string {
	return randomID("var")
}

func (RandomIDGenerator) NewImageID() string {
	return randomID("img")
}

func (RandomIDGenerator) NewInventoryReservationID() string {
	return randomID("res")
}

func (RandomIDGenerator) NewInventorySnapshotID() string {
	return randomID("inv_snap")
}

func (RandomIDGenerator) NewProductEventID() string {
	return randomID("evt")
}

type SystemClock struct{}

func (SystemClock) Now() time.Time {
	return time.Now().UTC()
}

func randomID(prefix string) string {
	var bytes [12]byte
	if _, err := rand.Read(bytes[:]); err != nil {
		return fmt.Sprintf("%s_%d", prefix, time.Now().UTC().UnixNano())
	}
	return prefix + "_" + hex.EncodeToString(bytes[:])
}
