package types

type PhraseType string

const (
	Welcome           PhraseType = "welcome"
	Goodbye           PhraseType = "goodbye"
	MembershipWarning PhraseType = "membership_warning"
	Info              PhraseType = "info"
)

type ContentSourceType string

const (
	Audio ContentSourceType = "audio"
)
