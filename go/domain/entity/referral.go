package entity

import (
	"time"

	"github.com/uptrace/bun"
)

const (
	UserReferralRewardLimit = 3000
	UserReferralRewardRate  = 5

	DriverReferralRewardLimit = 3000
	DriverReferralRewardRate  = 5
)

type UserReferral struct {
	bun.BaseModel `bun:"table:user_referral"`

	FromUserId    string    `bun:"from_user_id,pk"`
	ToUserId      string    `bun:"to_user_id"`
	RewardRate    int       `bun:"reward_rate"`
	CurrentReward int       `bun:"current_reward"`
	RewardLimit   int       `bun:"reward_limit"`
	CreateTime    time.Time `bun:"create_time"`
}

func (u *UserReferral) UseReward(price int) int {
	if price < 0 {
		return 0
	}

	rewardCandidate := price * u.RewardRate / 100
	if rewardCandidate+u.CurrentReward > u.RewardLimit {
		rewardCandidate = u.RewardLimit - u.CurrentReward
	}
	u.CurrentReward += rewardCandidate
	return rewardCandidate
}

type DriverReferral struct {
	bun.BaseModel `bun:"table:driver_referral"`

	FromDriverId  string    `bun:"from_driver_id,pk"`
	ToDriverId    string    `bun:"to_driver_id"`
	RewardRate    int       `bun:"reward_rate"`
	CurrentReward int       `bun:"current_reward"`
	RewardLimit   int       `bun:"reward_limit"`
	CreateTime    time.Time `bun:"create_time"`
}

func (d *DriverReferral) UseReward(price int) int {
	if price < 0 {
		return 0
	}

	rewardCandidate := price * d.RewardRate / 100
	if rewardCandidate+d.CurrentReward > d.RewardLimit {
		rewardCandidate = d.RewardLimit - d.CurrentReward
	}
	d.CurrentReward += rewardCandidate
	return rewardCandidate
}
