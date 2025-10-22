package main

import (
	"time"
)

type Mastery int

func (m Mastery) String() string {
	levels := []string{
		"",
		"+",
		"++",
		"+++",
		"++++",
		"+++++",
	}
	if int(m) < 0 || int(m) >= len(levels) {
		return ""
	}
	return levels[m]
}

func mastery(timesDone int, lastDone time.Time) Mastery {
	if timesDone <= 0 {
		return 0
	}

	var base int
	switch {
	case timesDone <= 2:
		base = 1
	case timesDone <= 5:
		base = 2
	case timesDone <= 9:
		base = 3
	case timesDone <= 14:
		base = 4
	default:
		base = 5
	}

	// decay by recency
	var decay int
	daysAgo := int(time.Since(lastDone).Hours() / 24)
	switch {
	case daysAgo <= 3:
		decay = 0
	case daysAgo <= 7:
		decay = 1
	case daysAgo <= 14:
		decay = 2
	default:
		decay = 3
	}

	level := base - decay
	if level < 0 {
		level = 0
	}

	return Mastery(level)
}
