package app

import (
	"crypto/rand"
	"fmt"
	"time"

	"github.com/srochagomes/mb-avaliacao/internal/domain"
)

const Instrument = "BTC-BRL"

type Service struct {
	store Transactor
	now   func() time.Time
	newID func() string
}

func New(store Transactor) *Service {
	return &Service{store: store, now: time.Now, newID: newUUID}
}

func newUUID() string {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		panic(err)
	}
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
}

func assetOf(side domain.Side) (domain.Asset, error) {
	switch side {
	case domain.SideBuy:
		return domain.AssetBRL, nil
	case domain.SideSell:
		return domain.AssetBTC, nil
	default:
		return "", ValidationError{Message: "invalid side"}
	}
}

func parseAsset(s string) (domain.Asset, error) {
	switch domain.Asset(s) {
	case domain.AssetBTC, domain.AssetBRL:
		return domain.Asset(s), nil
	default:
		return "", ValidationError{Message: "invalid asset"}
	}
}
