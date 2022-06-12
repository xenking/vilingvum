package domain

import "time"

const (
	StartTopicID int64 = 1
	DemoTopicID  int64 = 16
)

const (
	UserSubscriptionDuration = 180 * 24 * time.Hour
	TopicDeleteDelay         = 5 * time.Second
)
