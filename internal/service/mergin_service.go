package service

import "oms-contract/internal/domain"

type MarginService struct{}

func (m *MarginService) Freeze(o *domain.Order) error { return nil }
func (m *MarginService) Settle(o *domain.Order)       {}
