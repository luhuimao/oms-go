package service

import "oms-contract/internal/domain"

type RiskService struct{}

func (r *RiskService) Check(o *domain.Order) error {
	// demo: 永远通过
	return nil
}
