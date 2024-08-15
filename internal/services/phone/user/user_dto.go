package user

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type readDto struct {
	callingPlans       []readDtoCallingPlan
	costCenter         types.String
	department         types.String
	email              types.String
	emergencyAddressID types.String
	extensionID        types.String
	extensionNumber    types.Int64
	zoomUserID         types.String
	phoneNumbers       []readDtoPhoneNumber
	phoneUserID        types.String
	policy             *readDtoPolicy
	siteID             types.String
	status             types.String
}

type readDtoCallingPlan struct {
	callingPlanType    types.Int32
	billingAccountID   types.String
	billingAccountName types.String
}

type readDtoPhoneNumber struct {
	phoneNumberID types.String
	phoneNumber   types.String
}

type readDtoPolicy struct {
	adHocCallRecording                 *readDtoPolicyAdHocCallRecording
	adHocCallRecordingAccessMembers    []readDtoPolicyCallRecordingAccessMember
	autoCallRecording                  *readDtoPolicyAutoCallRecording
	autoCallRecordingAccessMembers     []readDtoPolicyCallRecordingAccessMember
	callOverflow                       *readDtoPolicyCallOverflow
	callPark                           *readDtoPolicyCallPark
	callTransferring                   *readDtoPolicyCallTransferring
	delegation                         types.Bool
	elevateToMeeting                   types.Bool
	emergencyAddressManagement         *readDtoPolicyEmergencyAddressManagement
	emergencyCallsToPsap               types.Bool
	callHandlingForwardingToOtherUsers *readDtoPolicyCallHandlingForwardingToOtherUsers
	handOffToRoom                      *readDtoPolicyHandOffToRoom
	internationalCalling               types.Bool
	mobileSwitchToCarrier              *readDtoPolicyMobileSwitchToCarrier
	selectOutboundCallerID             *readDtoPolicySelectOutboundCallerID
	sms                                *readDtoPolicySMS
	voicemail                          *readDtoPolicyVoicemail
	voicemailAccessMembers             []readDtoPolicyVoicemailAccessMember
	zoomPhoneOnMobile                  *readDtoPolicyZoomPhoneOnMobile
	personalAudioLibrary               *readDtoPolicyPersonalAudioLibrary
	voicemailTranscription             *readDtoPolicyVoicemailTranscription
	voicemailNotificationByEmail       *readDtoPolicyVoicemailNotificationByEmail
	sharedVoicemailNotificationByEmail *readDtoPolicySharedVoicemailNotificationByEmail
	checkVoicemailsOverPhone           *readDtoPolicyCheckVoicemailsOverPhone
	audioIntercom                      *readDtoPolicyAudioIntercom
	peerToPeerMedia                    *readDtoPolicyPeerToPeerMedia
	e2eEncryption                      *readDtoPolicyE2eEncryption
	outboundCalling                    *readDtoPolicyOutboundCalling
	outboundSMS                        *readDtoPolicyOutboundSMS
	allowEndUserEditCallHandling       *readDtoPolicyAllowEndUserEditCallHandling
}

type readDtoPolicyAdHocCallRecording struct {
	enable                 types.Bool
	locked                 types.Bool
	lockedBy               types.String
	playRecordingBeepTone  *readDtoPolicyPlayRecordingBeepTone
	recordingStartPrompt   types.Bool
	recordingTranscription types.Bool
}

type readDtoPolicyPlayRecordingBeepTone struct {
	enable               types.Bool
	playBeepVolume       types.Int32
	playBeepTimeInterval types.Int32
	playBeepMember       types.String
}

type readDtoPolicyCallRecordingAccessMember struct {
	accessUserID  types.String
	allowDelete   types.Bool
	allowDownload types.Bool
	sharedID      types.String
}

type readDtoPolicyAutoCallRecording struct {
	allowStopResumeRecording     types.Bool
	disconnectOnRecordingFailure types.Bool
	enable                       types.Bool
	locked                       types.Bool
	lockedBy                     types.String
	recordingCalls               types.String
	recordingExplicitConsent     types.Bool
	recordingStartPrompt         types.Bool
	recordingTranscription       types.Bool
	playRecordingBeepTone        *readDtoPolicyPlayRecordingBeepTone
}

type readDtoPolicyCallOverflow struct {
	callOverflowType types.Int32
	enable           types.Bool
	locked           types.Bool
	lockedBy         types.String
	modified         types.Bool
}

type readDtoPolicyCallPark struct {
	callNotPickedUpAction types.Int32
	enable                types.Bool
	expirationPeriod      types.Int32
	forwardTo             *readDtoPolicyCallParkForwardTo
	locked                types.Bool
	lockedBy              types.String
}

type readDtoPolicyCallParkForwardTo struct {
	displayName     types.String
	extensionID     types.String
	extensionNumber types.Int64
	extensionType   types.String
	forwardToID     types.String
}

type readDtoPolicyCallTransferring struct {
	callTransferringType types.Int32
	enable               types.Bool
	locked               types.Bool
	lockedBy             types.String
}

type readDtoPolicyEmergencyAddressManagement struct {
	enable               types.Bool
	promptDefaultAddress types.Bool
}

type readDtoPolicyCallHandlingForwardingToOtherUsers struct {
	enable             types.Bool
	callForwardingType types.Int32
	locked             types.Bool
	lockedBy           types.String
	modified           types.Bool
}

type readDtoPolicyHandOffToRoom struct {
	enable   types.Bool
	locked   types.Bool
	lockedBy types.String
}

type readDtoPolicyMobileSwitchToCarrier struct {
	enable   types.Bool
	locked   types.Bool
	lockedBy types.String
}

type readDtoPolicySelectOutboundCallerID struct {
	enable                    types.Bool
	allowHideOutboundCallerID types.Bool
	locked                    types.Bool
	lockedBy                  types.String
}

type readDtoPolicySMS struct {
	enable                    types.Bool
	internationalSMS          types.Bool
	InternationalSMSCountries []types.String
	locked                    types.Bool
	lockedBy                  types.String
}

type readDtoPolicyVoicemail struct {
	allowDelete        types.Bool
	allowDownload      types.Bool
	allowTranscription types.Bool
	allowVideomail     types.Bool
	enable             types.Bool
}

type readDtoPolicyVoicemailAccessMember struct {
	accessUserID  types.String
	allowDelete   types.Bool
	allowDownload types.Bool
	allowSharing  types.Bool
	sharedID      types.String
}

type readDtoPolicyZoomPhoneOnMobile struct {
	allowCallingSMSMms types.Bool
	enable             types.Bool
	locked             types.Bool
	lockedBy           types.String
}

type readDtoPolicyPersonalAudioLibrary struct {
	enable                                        types.Bool
	locked                                        types.Bool
	lockedBy                                      types.String
	modified                                      types.Bool
	allowMusicOnHoldCustomization                 types.Bool
	allowVoicemailAndMessageGreetingCustomization types.Bool
}

type readDtoPolicyVoicemailTranscription struct {
	enable   types.Bool
	locked   types.Bool
	lockedBy types.String
	modified types.Bool
}

type readDtoPolicyVoicemailNotificationByEmail struct {
	includeVoicemailFile          types.Bool
	includeVoicemailTranscription types.Bool
	enable                        types.Bool
	locked                        types.Bool
	lockedBy                      types.String
	modified                      types.Bool
}

type readDtoPolicySharedVoicemailNotificationByEmail struct {
	enable   types.Bool
	locked   types.Bool
	lockedBy types.String
	modified types.Bool
}

type readDtoPolicyCheckVoicemailsOverPhone struct {
	enable   types.Bool
	locked   types.Bool
	lockedBy types.String
	modified types.Bool
}

type readDtoPolicyAudioIntercom struct {
	enable   types.Bool
	locked   types.Bool
	lockedBy types.String
	modified types.Bool
}

type readDtoPolicyPeerToPeerMedia struct {
	enable   types.Bool
	locked   types.Bool
	lockedBy types.String
	modified types.Bool
}

type readDtoPolicyE2eEncryption struct {
	enable   types.Bool
	locked   types.Bool
	lockedBy types.String
	modified types.Bool
}

type readDtoPolicyOutboundCalling struct {
	enable   types.Bool
	locked   types.Bool
	modified types.Bool
}

type readDtoPolicyOutboundSMS struct {
	enable   types.Bool
	locked   types.Bool
	modified types.Bool
}

type readDtoPolicyAllowEndUserEditCallHandling struct {
	enable   types.Bool
	locked   types.Bool
	modified types.Bool
}

type createDto struct {
	email           types.String
	firstName       types.String
	lastName        types.String
	extensionNumber types.String
	// Experimental: For tring to call user api
	zoomUserID types.String
}

type createdDto struct {
	zoomUserID types.String
}

type updateDto struct {
	zoomUserID         types.String
	emergencyAddressID types.String
	extensionNumber    types.String
	policy             *updateDtoPolicy
	siteID             types.String
	templateID         types.String
}

type updateDtoPolicy struct {
	adHocCallRecording                 *updateDtoPolicyAdHocCallRecording
	autoCallRecording                  *updateDtoPolicyAutoCallRecording
	callOverflow                       *updateDtoPolicyCallOverflow
	callPark                           *updateDtoPolicyCallPark
	callTransferring                   *updateDtoPolicyCallTransferring
	delegation                         types.Bool
	elevateToMeeting                   types.Bool
	emergencyAddressManagement         *updateDtoPolicyEmergencyAddressManagement
	emergencyCallsToPsap               types.Bool
	forwardingToExternalNumbers        types.Bool
	callHandlingForwardingToOtherUsers *updateDtoPolicyCallHandlingForwardingToOtherUsers
	handOffToRoom                      *updateDtoPolicyHandOffToRoom
	internationalCalling               types.Bool
	mobileSwitchToCarrier              *updateDtoPolicyMobileSwitchToCarrier
	selectOutboundCallerID             *updateDtoPolicySelectOutboundCallerID
	sms                                *updateDtoPolicySMS
	voicemail                          *updateDtoPolicyVoicemail
	zoomPhoneOnMobile                  *updateDtoPolicyZoomPhoneOnMobile
	personalAudioLibrary               *updateDtoPolicyPersonalAudioLibrary
	voicemailTranscription             *updateDtoPolicyVoicemailTranscription
	voicemailNotificationByEmail       *updateDtoPolicyVoicemailNotificationByEmail
	sharedVoicemailNotificationByEmail *updateDtoPolicySharedVoicemailNotificationByEmail
	checkVoicemailsOverPhone           *updateDtoPolicyCheckVoicemailsOverPhone
	audioIntercom                      *updateDtoPolicyAudioIntercom
	e2eEncryption                      *updateDtoPolicyE2EEncryption
}

type updateDtoPolicyPlayRecordingBeepTone struct {
	enable               types.Bool
	playBeepVolume       types.Int32
	playBeepTimeInterval types.Int32
	playBeepMember       types.String
}

type updateDtoPolicyAdHocCallRecording struct {
	enable                 types.Bool
	recordingStartPrompt   types.Bool
	recordingTranscription types.Bool
	playRecordingBeepTone  *updateDtoPolicyPlayRecordingBeepTone
}

type updateDtoPolicyAutoCallRecording struct {
	allowStopResumeRecording     types.Bool
	disconnectOnRecordingFailure types.Bool
	enable                       types.Bool
	recordingCalls               types.String
	recordingExplicitConsent     types.Bool
	recordingStartPrompt         types.Bool
	recordingTranscription       types.Bool
	playRecordingBeepTone        *updateDtoPolicyPlayRecordingBeepTone
}

type updateDtoPolicyCallOverflow struct {
	callOverflowType types.Int32
	enable           types.Bool
}

type updateDtoPolicyCallPark struct {
	callNotPickedUpAction types.Int32
	enable                types.Bool
	expirationPeriod      types.Int32
	forwardToExtensionID  types.String
}

type updateDtoPolicyCallTransferring struct {
	callTransferringType types.Int32
	enable               types.Bool
}

type updateDtoPolicyEmergencyAddressManagement struct {
	enable               types.Bool
	promptDefaultAddress types.Bool
}

type updateDtoPolicyCallHandlingForwardingToOtherUsers struct {
	enable             types.Bool
	callForwardingType types.Int32
}

type updateDtoPolicyHandOffToRoom struct {
	enable types.Bool
}

type updateDtoPolicyMobileSwitchToCarrier struct {
	enable types.Bool
}

type updateDtoPolicySelectOutboundCallerID struct {
	enable                    types.Bool
	allowHideOutboundCallerID types.Bool
}

type updateDtoPolicySMS struct {
	enable                    types.Bool
	internationalSMS          types.Bool
	InternationalSMSCountries []types.String
}

type updateDtoPolicyVoicemail struct {
	allowDelete        types.Bool
	allowDownload      types.Bool
	allowTranscription types.Bool
	allowVideomail     types.Bool
	enable             types.Bool
}

type updateDtoPolicyZoomPhoneOnMobile struct {
	allowCallingSMSMms types.Bool
	enable             types.Bool
}

type updateDtoPolicyPersonalAudioLibrary struct {
	allowMusicOnHoldCustomization                 types.Bool
	allowVoicemailAndMessageGreetingCustomization types.Bool
	enable                                        types.Bool
}

type updateDtoPolicyVoicemailTranscription struct {
	enable types.Bool
}

type updateDtoPolicyVoicemailNotificationByEmail struct {
	includeVoicemailFile          types.Bool
	includeVoicemailTranscription types.Bool
	enable                        types.Bool
}

type updateDtoPolicySharedVoicemailNotificationByEmail struct {
	enable types.Bool
}

type updateDtoPolicyCheckVoicemailsOverPhone struct {
	enable types.Bool
}

type updateDtoPolicyAudioIntercom struct {
	enable types.Bool
}

type updateDtoPolicyE2EEncryption struct {
	enable types.Bool
}
