package autoreceptionistivr

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type readDto struct {
	autoReceptionistID   types.String
	hoursType            types.String
	holidayID            types.String
	audioPrompt          *readDtoAudioPrompt
	callerEntersNoAction *readDtoCallerEntersNoAction
	keyActions           []*readDtoKeyAction
}

type readDtoAudioPrompt struct {
	id   types.String
	name types.String
}

type readDtoCallerEntersNoAction struct {
	action            types.Int32
	audioPromptRepeat types.Int32
	forwardTo         *readDtoCallerEntersNoActionForwardTo
}

type readDtoCallerEntersNoActionForwardTo struct {
	displayName     types.String
	extensionID     types.String
	extensionNumber types.String
	id              types.String
}

type readDtoKeyAction struct {
	action            types.Int32
	key               types.String
	target            *readDtoKeyActionTarget
	voicemailGreeting *readDtoKeyActionVoicemailGreeting
}

type readDtoKeyActionTarget struct {
	displayName     types.String
	extensionID     types.String
	extensionNumber types.String
	id              types.String
	phoneNumber     types.String
}

type readDtoKeyActionVoicemailGreeting struct {
	id   types.String
	name types.String
}

type updateDto struct {
	autoReceptionistID   types.String
	holidayID            types.String
	hoursType            types.String
	audioPromptID        types.String
	callerEntersNoAction *updateDtoCallerEntersNoAction
	keyActions           []*updateDtoKeyAction
}

type updateDtoCallerEntersNoAction struct {
	action               types.Int32
	auditPromptRepeat    types.Int32
	forwardToExtensionID types.String
}

type updateDtoKeyAction struct {
	key                 types.String
	action              types.Int32
	target              *updateDtoKeyActionTarget
	voiceMailGreetingId types.String
}

type updateDtoKeyActionTarget struct {
	extensionID types.String
	phoneNumber types.String
}
