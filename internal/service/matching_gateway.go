package service

import "oms-contract/internal/domain"

type MatchingGateway interface {
	SendLiquidationOrder(order *domain.LiquidationOrder) error
}
