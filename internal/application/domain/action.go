package domain

type Action int8

const (
	ActionNone Action = iota
	ActionFeedback
	ActionTestReport
)
