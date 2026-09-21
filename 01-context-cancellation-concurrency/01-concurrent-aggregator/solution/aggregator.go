package aggregator

import (
	"context"
	"log"
	"time"

	"github.com/medunes/go-kata/01-context-cancellation-concurrency/01-concurrent-aggregator/order"
	"github.com/medunes/go-kata/01-context-cancellation-concurrency/01-concurrent-aggregator/profile"
)

type UserAggregator struct {
	ProfileService  profile.Service
	OrdersService   order.Service
	Context         context.Context
	TimeoutDuration int
	Logger          log.Logger
}

type Option func(userAgg *UserAggregator)

func WithLogger(logger log.Logger) {

}

func WithTimeout(int) {

}

func NewUserAggregator(
	profileService profile.Service,
	ordersService order.Service,
	opts ...Option,
) *UserAggregator {
	return &UserAggregator{
		ProfileService:  profileService,
		OrdersService:   ordersService,
		Context:         context,
		TimeoutDuration: timeoutDuration * int(time.Second),
		Logger:          logger,
	}
}

func (agg UserAggregator) Aggregate(id int) {
	userId, _ := agg.ProfileService.Get(agg.Context, id)
}
