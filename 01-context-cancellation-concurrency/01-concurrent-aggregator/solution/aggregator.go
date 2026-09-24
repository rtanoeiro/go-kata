package aggregator

import (
	"context"
	"log/slog"
	"time"

	"github.com/rtanoeiro/go-kata/01-context-cancellation-concurrency/01-concurrent-aggregator/order"
	"github.com/rtanoeiro/go-kata/01-context-cancellation-concurrency/01-concurrent-aggregator/profile"
	"golang.org/x/sync/errgroup"
)

type UserAggregator struct {
	ProfileService  profile.Service
	OrdersService   order.Service
	TimeoutDuration time.Duration
	Logger          *slog.Logger
}

type Option func(userAgg *UserAggregator)

func WithLogger(logger *slog.Logger) Option {
	return func(userAgg *UserAggregator) {
		userAgg.Logger = logger
	}
}

func WithTimeout(timeoutDuration time.Duration) Option {
	return func(userAgg *UserAggregator) {
		userAgg.TimeoutDuration = timeoutDuration
	}

}

func NewUserAggregator(
	ordersService order.Service,
	profileService profile.Service,
	opts ...Option,
) *UserAggregator {
	userAgg := &UserAggregator{}
	userAgg.OrdersService = ordersService
	userAgg.ProfileService = profileService

	for _, opt := range opts {
		opt(userAgg)
	}
	return userAgg
}

type AggregatedProfile struct {
	Name string
	Cost float64
}

func (agg UserAggregator) Aggregate(ctx context.Context, id int) ([]*AggregatedProfile, error) {
	localCtx, cancelCtx := context.WithCancel(ctx)
	// If the Timeout is equal to 0, then context is cancelled instantly, not giving time
	// for further operations to complete
	if agg.TimeoutDuration > 0 {
		localCtx, cancelCtx = context.WithTimeout(ctx, agg.TimeoutDuration)
	}
	defer cancelCtx()

	errGroup := errgroup.Group{}

	var aggProfile []*AggregatedProfile
	var userProfile *profile.Profile
	var userOrders []*order.Order
	var errProfile, errOrders error

	errGroup.Go(
		func() error {
			userProfile, errProfile = agg.ProfileService.Get(localCtx, id)
			if errProfile != nil {
				agg.Logger.Info("Failed to get user profile", "error", errProfile)
				return errProfile
			}
			return nil
		})
	errGroup.Go(
		func() error {
			userOrders, errOrders = agg.OrdersService.GetAll(localCtx, id)
			if errOrders != nil {
				agg.Logger.Info("Failed to get user orders", "error", errOrders)
				return errOrders
			}
			return nil
		})

	if err := errGroup.Wait(); err != nil {
		agg.Logger.Info("Failed to get data, returning empty profile", "error", err)
		return nil, err
	}

	for _, order := range userOrders {
		agg.Logger.Info("Adding order to Aggregated Profile", "order", order)
		aggProfile = append(aggProfile, &AggregatedProfile{
			Name: userProfile.Name,
			Cost: order.Cost,
		})
	}
	agg.Logger.Info("aggregation complete successfully", "user_id", userProfile.Id)

	return aggProfile, nil
}
