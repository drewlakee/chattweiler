package model

type PhraseType string

const (
	WelcomeType          PhraseType = "welcome"
	GoodbyeType          PhraseType = "goodbye"
	MembershipWarninType PhraseType = "membership_warning"
	InfoType             PhraseType = "info"
	AudioRequestType     PhraseType = "audio_request"
	PictureRequestType   PhraseType = "picture_request"
)

type ContentSourceType string

const (
	AudioType   ContentSourceType = "audio"
	PictureType ContentSourceType = "picture"
)
