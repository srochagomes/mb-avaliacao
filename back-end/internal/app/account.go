package app

import (
	"context"
	"strings"
)

func (s *Service) CreateAccount(ctx context.Context, label string) (Account, error) {
	label = strings.TrimSpace(label)
	if label == "" {
		return Account{}, ValidationError{Message: "label required"}
	}
	if len(label) > 40 {
		return Account{}, ValidationError{Message: "label too long"}
	}
	var created Account
	err := s.store.Within(ctx, func(tx Tx) error {
		created = tx.CreateAccount(label, s.now().UTC(), s.newID())
		return nil
	})
	return created, err
}

func (s *Service) ListAccounts(ctx context.Context) ([]Account, error) {
	var accounts []Account
	err := s.store.Within(ctx, func(tx Tx) error {
		accounts = tx.Accounts()
		return nil
	})
	return accounts, err
}

func (s *Service) requireAccount(tx Tx, id string) error {
	if strings.TrimSpace(id) == "" {
		return ValidationError{Message: "account_id required"}
	}
	_, err := tx.Account(id)
	return err
}
