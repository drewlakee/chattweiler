package model

type PhraseType string

const (
	WelcomeType           PhraseType = "welcome"
	GoodbyeType           PhraseType = "goodbye"
	MembershipWarningType PhraseType = "membership_warning"
	InfoType              PhraseType = "info"
	ContentRequestType    PhraseType = "content_request"
)

type MediaContentType string

const (
	AudioType   MediaContentType = "audio"
	PictureType MediaContentType = "picture"
)
