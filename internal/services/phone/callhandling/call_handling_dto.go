package callhandling

import (
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type readDtoBusinessHours struct {
	extensionID    types.String
	customHours    *readDtoBusinessHoursCustomHours
	callHandling   *readDtoBusinessHoursCallHandling
	callForwarding *readDtoHolidayHoursCallForwarding
}

type readDtoBusinessHoursCustomHours struct {
	typ                 types.Int32
	allowMembersToReset types.Bool
	settings            []*readDtoBusinessHoursCustomHoursSetting
}

type readDtoBusinessHoursCustomHoursSetting struct {
	weekday types.Int32
	typ     types.Int32
	from    types.String
	to      types.String
}

type readDtoBusinessHoursCallHandling struct {
	callNotAnswerAction                     types.Int32
	forwardToExtensionID                    types.String
	busyOnAnotherCallAction                 types.Int32
	busyForwardToExtensionID                types.String
	allowCallersCheckVoicemail              types.Bool
	allowMembersToReset                     types.Bool
	audioWhileConnectingID                  types.String
	callDistribution                        *readDtoBusinessHoursCallHandlingCallDistribution
	busyRequirePress1BeforeConnecting       types.Bool
	unAnsweredRequirePress1BeforeConnecting types.Bool
	overflowPlayCalleeVoicemailGreeting     types.Bool
	playCalleeVoicemailGreeting             types.Bool
	busyPlayCalleeVoicemailGreeting         types.Bool
	phoneNumber                             types.String
	phoneNumberDescription                  types.String
	busyPhoneNumber                         types.String
	busyPhoneNumberDescription              types.String
	connectToOperator                       types.Bool
	greetingPromptID                        types.String
	maxCallInQueue                          types.Int32
	maxWaitTime                             types.Int32
	musicOnHoldID                           types.String
	operatorExtensionID                     types.String
	receiveCall                             types.Bool
	ringMode                                types.String
	voiceMailGreetingID                     types.String
	wrapUpTime                              types.Int32
}

type readDtoBusinessHoursCallHandlingCallDistribution struct {
	handleMultipleCalls          types.Bool
	ringDuration                 types.Int32
	ringMode                     types.String
	skipOfflineDevicePhoneNumber types.Bool
}

type readDtoClosedHours struct {
	extensionID    types.String
	callHandling   *readDtoClosedHoursCallHandling
	callForwarding *readDtoHolidayHoursCallForwarding
}

type readDtoClosedHoursCallHandling struct {
	callNotAnswerAction                     types.Int32
	forwardToExtensionID                    types.String
	busyOnAnotherCallAction                 types.Int32
	busyForwardToExtensionID                types.String
	allowCallersCheckVoicemail              types.Bool
	busyRequirePress1BeforeConnecting       types.Bool
	unAnsweredRequirePress1BeforeConnecting types.Bool
	overflowPlayCalleeVoicemailGreeting     types.Bool
	playCalleeVoicemailGreeting             types.Bool
	busyPlayCalleeVoicemailGreeting         types.Bool
	phoneNumber                             types.String
	phoneNumberDescription                  types.String
	busyPhoneNumber                         types.String
	busyPhoneNumberDescription              types.String
	connectToOperator                       types.Bool
	greetingPromptID                        types.String
	maxWaitTime                             types.Int32
	operatorExtensionID                     types.String
	ringMode                                types.String
}

type readDtoHolidayHours struct {
	extensionID    types.String
	holidayID      types.String
	holiday        *readDtoHolidayHoursHoliday
	callHandling   *readDtoHolidayHoursCallHandling
	callForwarding *readDtoHolidayHoursCallForwarding
}

type readDtoHolidayHoursHoliday struct {
	name types.String
	from timetypes.RFC3339
	to   timetypes.RFC3339
}

type readDtoHolidayHoursCallHandling struct {
	callNotAnswerAction                     types.Int32
	forwardToExtensionID                    types.String
	allowCallersCheckVoicemail              types.Bool
	unAnsweredRequirePress1BeforeConnecting types.Bool
	overflowPlayCalleeVoicemailGreeting     types.Bool
	playCalleeVoicemailGreeting             types.Bool
	phoneNumber                             types.String
	phoneNumberDescription                  types.String
	connectToOperator                       types.Bool
	maxWaitTime                             types.Int32
	operatorExtensionID                     types.String
	ringMode                                types.String
}

type readDtoHolidayHoursCallForwarding struct {
	requirePress1BeforeConnecting types.Bool
	enableZoomMobileApps          types.Bool
	enableZoomDesktopApps         types.Bool
	enableZoomPhoneApplianceApps  types.Bool
	settings                      []*readDtoCallForwardingSetting
}

type readDtoCallForwardingSetting struct {
	id              types.String
	description     types.String
	enable          types.Bool
	phoneNumber     types.String
	externalContact *readDtoCallForwardingSettingsExternalContact
}

type readDtoCallForwardingSettingsExternalContact struct {
	externalContactID types.String
}

type createHolidayDto struct {
	extensionID types.String
	settingType settingType
	name        types.String
	from        timetypes.RFC3339
	to          timetypes.RFC3339
}

type createdHolidayDto struct {
	holidayID types.String
}

type createCallForwardingDto struct {
	extensionID types.String
	settingType settingType
	holidayID   types.String
	description types.String
	phoneNumber types.String
}

type createdCallForwardingDto struct {
	callForwardingID types.String
}

type settingType string

const (
	settingTypeBusinessHours settingType = "business_hours"
	settingTypeClosedHours   settingType = "closed_hours"
	settingTypeHolidayHours  settingType = "holiday_hours"
)

type patchCustomHoursDto struct {
	extensionID         types.String
	settingType         settingType
	typ                 types.Int32
	allowMembersToReset types.Bool
	settings            []*patchCustomHoursDtoSetting
}

type patchCustomHoursDtoSetting struct {
	weekday types.Int32
	typ     types.Int32
	from    types.String
	to      types.String
}

type patchCallHandlingDto struct {
	extensionID types.String
	settingType settingType
	settings    *patchCallHandlingDtoSettings
}

type patchCallHandlingDtoSettings struct {
	holidayID                               types.String
	callNotAnswerAction                     types.Int32
	forwardToExtensionID                    types.String
	busyOnAnotherCallAction                 types.Int32
	busyForwardToExtensionID                types.String
	allowCallersCheckVoicemail              types.Bool
	allowMembersToReset                     types.Bool
	audioWhileConnectingID                  types.String
	callDistribution                        *patchCallHandlingDtoSettingsDistribution
	busyRequirePress1BeforeConnecting       types.Bool
	unAnsweredRequirePress1BeforeConnecting types.Bool
	overflowPlayCalleeVoicemailGreeting     types.Bool
	playCalleeVoicemailGreeting             types.Bool
	busyPlayCalleeVoicemailGreeting         types.Bool
	phoneNumber                             types.String
	phoneNumberDescription                  types.String
	busyPhoneNumber                         types.String
	busyPhoneNumberDescription              types.String
	connectToOperator                       types.Bool
	greetingPromptID                        types.String
	maxCallInQueue                          types.Int32
	maxWaitTime                             types.Int32
	musicOnHoldID                           types.String
	operatorExtensionID                     types.String
	receiveCall                             types.Bool
	ringMode                                types.String
	voiceMailGreetingID                     types.String
	wrapUpTime                              types.Int32
}

type patchCallHandlingDtoSettingsDistribution struct {
	handleMultipleCalls          types.Bool
	ringDuration                 types.Int32
	ringMode                     types.String
	skipOfflineDevicePhoneNumber types.Bool
}

type patchHolidayDto struct {
	extensionID types.String
	settingType settingType
	holidayID   types.String
	name        types.String
	from        timetypes.RFC3339
	to          timetypes.RFC3339
}

type patchCallForwardingDto struct {
	extensionID                   types.String
	holidayID                     types.String
	settingType                   settingType
	requirePress1BeforeConnecting types.Bool
	enableZoomMobileApps          types.Bool
	enableZoomDesktopApps         types.Bool
	enableZoomPhoneApplianceApps  types.Bool
	settings                      []*patchCallForwardingDtoSetting
}

type patchCallForwardingDtoSetting struct {
	id                types.String
	description       types.String
	enable            types.Bool
	phoneNumber       types.String
	externalContactID types.String
}

type deleteCallForwardingDto struct {
	extensionID      types.String
	settingType      settingType
	callForwardingID types.String
}

type deleteHolidayDto struct {
	extensionID types.String
	settingType settingType
	holidayID   types.String
}
