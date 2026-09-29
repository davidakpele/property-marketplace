package domain

import "errors"

var (
	ErrListingNotFound    = errors.New("listing not found")
	ErrAgentNotFound      = errors.New("agent not found")
	ErrInvalidListingType = errors.New("invalid listing type")
)
