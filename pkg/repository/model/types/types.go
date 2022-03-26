package types

type PhraseType string

const (
	Welcome           PhraseType = "welcome"
	Goodbye           PhraseType = "goodbye"
	MembershipWarning PhraseType = "membership_warning"
	Info              PhraseType = "info"
	AudioRequest      PhraseType = "audio_request"
	PictureRequest    PhraseType = "picture_request"
)

type ContentSourceType string

const (
	Audio   ContentSourceType = "audio"
	Picture ContentSourceType = "picture"
)
