package workflow

import "errors"

var (
	ErrSlugEmpty     = errors.New("slug cannot be empty")
	ErrVerifyMissing = errors.New("verify.md not found - run verify before archiving")
)
