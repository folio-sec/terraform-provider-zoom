package user

import (
	"context"
	"errors"
	"fmt"

	"github.com/folio-sec/terraform-provider-zoom/generated/api/zoomphone"
	"github.com/folio-sec/terraform-provider-zoom/generated/api/zoomuser"
	"github.com/folio-sec/terraform-provider-zoom/internal/util"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/samber/lo"
)

func newUserCrud(phoneClient *zoomphone.Client, userClient *zoomuser.Client) *userCrud {
	return &userCrud{
		phoneClient: phoneClient,
		userClient:  userClient,
	}
}

type userCrud struct {
	phoneClient *zoomphone.Client
	userClient  *zoomuser.Client
}

func (c *userCrud) read(ctx context.Context, zoomUserID string) (*readDto, error) {
	detail, err := c.phoneClient.PhoneUser(ctx, zoomphone.PhoneUserParams{
		UserId: zoomUserID,
	})
	if err != nil {
		var status *zoomphone.ErrorResponseStatusCode
		if errors.As(err, &status) {
			if status.StatusCode == 400 && status.Response.Code.Value == 300 {
				return nil, nil // already deleted
			}
		}
		return nil, fmt.Errorf("unable to read phone user: %v", err)
	}

	var policy *readDtoPolicy
	if detail.Policy.IsSet() {
		policy = &readDtoPolicy{
			adHocCallRecording: lo.TernaryF(detail.Policy.Value.AdHocCallRecording.IsSet(), func() *readDtoPolicyAdHocCallRecording {
				v := detail.Policy.Value.AdHocCallRecording.Value
				return &readDtoPolicyAdHocCallRecording{
					enable:   util.FromOptBool(v.Enable),
					locked:   util.FromOptBool(v.Locked),
					lockedBy: util.FromOptString(v.LockedBy),
					playRecordingBeepTone: lo.Ternary(v.PlayRecordingBeepTone.IsSet(), &readDtoPolicyPlayRecordingBeepTone{
						enable:               util.FromOptBool(v.PlayRecordingBeepTone.Value.Enable),
						playBeepVolume:       util.FromOptInt(v.PlayRecordingBeepTone.Value.PlayBeepVolume),
						playBeepTimeInterval: util.FromOptInt(v.PlayRecordingBeepTone.Value.PlayBeepTimeInterval),
						playBeepMember:       util.FromOptString(v.PlayRecordingBeepTone.Value.PlayBeepMember),
					}, nil),
					recordingStartPrompt:   util.FromOptBool(v.RecordingStartPrompt),
					recordingTranscription: util.FromOptBool(v.RecordingTranscription),
				}
			}, lo.Nil),
			adHocCallRecordingAccessMembers: lo.Map(detail.Policy.Value.AdHocCallRecordingAccessMembers, func(v zoomphone.PhoneUserOKPolicyAdHocCallRecordingAccessMembersItem, _ int) readDtoPolicyCallRecordingAccessMember {
				return readDtoPolicyCallRecordingAccessMember{
					accessUserID:  util.FromOptString(v.AccessUserID),
					allowDelete:   util.FromOptBool(v.AllowDelete),
					allowDownload: util.FromOptBool(v.AllowDownload),
					sharedID:      util.FromOptString(v.SharedID),
				}
			}),
			autoCallRecording: lo.TernaryF(detail.Policy.Value.AutoCallRecording.IsSet(), func() *readDtoPolicyAutoCallRecording {
				v := detail.Policy.Value.AutoCallRecording.Value
				return &readDtoPolicyAutoCallRecording{
					allowStopResumeRecording:     util.FromOptBool(v.AllowStopResumeRecording),
					disconnectOnRecordingFailure: util.FromOptBool(v.DisconnectOnRecordingFailure),
					enable:                       util.FromOptBool(v.Enable),
					locked:                       util.FromOptBool(v.Locked),
					lockedBy:                     util.FromOptString(v.LockedBy),
					recordingCalls:               util.FromOptString(v.RecordingCalls),
					recordingExplicitConsent:     util.FromOptBool(v.RecordingExplicitConsent),
					recordingStartPrompt:         util.FromOptBool(v.RecordingStartPrompt),
					recordingTranscription:       util.FromOptBool(v.RecordingTranscription),
					playRecordingBeepTone: lo.Ternary(v.PlayRecordingBeepTone.IsSet(), &readDtoPolicyPlayRecordingBeepTone{
						enable:               util.FromOptBool(v.PlayRecordingBeepTone.Value.Enable),
						playBeepVolume:       util.FromOptInt(v.PlayRecordingBeepTone.Value.PlayBeepVolume),
						playBeepTimeInterval: util.FromOptInt(v.PlayRecordingBeepTone.Value.PlayBeepTimeInterval),
						playBeepMember:       util.FromOptString(v.PlayRecordingBeepTone.Value.PlayBeepMember),
					}, nil),
				}
			}, lo.Nil),
			autoCallRecordingAccessMembers: lo.Map(detail.Policy.Value.AutoCallRecordingAccessMembers, func(v zoomphone.PhoneUserOKPolicyAutoCallRecordingAccessMembersItem, _ int) readDtoPolicyCallRecordingAccessMember {
				return readDtoPolicyCallRecordingAccessMember{
					accessUserID:  util.FromOptString(v.AccessUserID),
					allowDelete:   util.FromOptBool(v.AllowDelete),
					allowDownload: util.FromOptBool(v.AllowDownload),
					sharedID:      util.FromOptString(v.SharedID),
				}
			}),
			callOverflow: lo.TernaryF(detail.Policy.Value.CallOverflow.IsSet(), func() *readDtoPolicyCallOverflow {
				v := detail.Policy.Value.CallOverflow.Value
				return &readDtoPolicyCallOverflow{
					callOverflowType: util.FromOptInt(v.CallOverflowType),
					enable:           util.FromOptBool(v.Enable),
					locked:           util.FromOptBool(v.Locked),
					lockedBy:         util.FromOptString(v.LockedBy),
					modified:         util.FromOptBool(v.Modified),
				}
			}, lo.Nil),
			callPark: lo.TernaryF(detail.Policy.Value.CallPark.IsSet(), func() *readDtoPolicyCallPark {
				v := detail.Policy.Value.CallPark.Value
				return &readDtoPolicyCallPark{
					callNotPickedUpAction: util.FromOptInt(v.CallNotPickedUpAction),
					enable:                util.FromOptBool(v.Enable),
					expirationPeriod:      util.FromOptInt(v.ExpirationPeriod),
					forwardTo: lo.Ternary(v.ForwardTo.IsSet(), &readDtoPolicyCallParkForwardTo{
						displayName:     util.FromOptString(v.ForwardTo.Value.DisplayName),
						extensionID:     util.FromOptString(v.ForwardTo.Value.ExtensionID),
						extensionNumber: util.FromOptInt64(v.ForwardTo.Value.ExtensionNumber),
						extensionType:   util.FromOptString(v.ForwardTo.Value.ExtensionType),
						forwardToID:     util.FromOptString(v.ForwardTo.Value.ID),
					}, nil),
					locked:   util.FromOptBool(v.Locked),
					lockedBy: util.FromOptString(v.LockedBy),
				}
			}, lo.Nil),
			callTransferring: lo.TernaryF(detail.Policy.Value.CallTransferring.IsSet(), func() *readDtoPolicyCallTransferring {
				v := detail.Policy.Value.CallTransferring.Value
				return &readDtoPolicyCallTransferring{
					callTransferringType: util.FromOptInt(v.CallTransferringType),
					enable:               util.FromOptBool(v.Enable),
					locked:               util.FromOptBool(v.Locked),
					lockedBy:             util.FromOptString(v.LockedBy),
				}
			}, lo.Nil),
			delegation:       util.FromOptBool(detail.Policy.Value.Delegation),
			elevateToMeeting: util.FromOptBool(detail.Policy.Value.ElevateToMeeting),
			emergencyAddressManagement: lo.TernaryF(detail.Policy.Value.EmergencyAddressManagement.IsSet(), func() *readDtoPolicyEmergencyAddressManagement {
				v := detail.Policy.Value.EmergencyAddressManagement.Value
				return &readDtoPolicyEmergencyAddressManagement{
					enable:               util.FromOptBool(v.Enable),
					promptDefaultAddress: util.FromOptBool(v.PromptDefaultAddress),
				}
			}, lo.Nil),
			emergencyCallsToPsap: util.FromOptBool(detail.Policy.Value.EmergencyCallsToPsap),
			callHandlingForwardingToOtherUsers: lo.TernaryF(detail.Policy.Value.CallHandlingForwardingToOtherUsers.IsSet(), func() *readDtoPolicyCallHandlingForwardingToOtherUsers {
				v := detail.Policy.Value.CallHandlingForwardingToOtherUsers.Value
				return &readDtoPolicyCallHandlingForwardingToOtherUsers{
					enable:             util.FromOptBool(v.Enable),
					callForwardingType: util.FromOptInt(v.CallForwardingType),
					locked:             util.FromOptBool(v.Locked),
					lockedBy:           util.FromOptString(v.LockedBy),
					modified:           util.FromOptBool(v.Modified),
				}
			}, lo.Nil),
			handOffToRoom: lo.TernaryF(detail.Policy.Value.HandOffToRoom.IsSet(), func() *readDtoPolicyHandOffToRoom {
				v := detail.Policy.Value.HandOffToRoom.Value
				return &readDtoPolicyHandOffToRoom{
					enable:   util.FromOptBool(v.Enable),
					locked:   util.FromOptBool(v.Locked),
					lockedBy: util.FromOptString(v.LockedBy),
				}
			}, lo.Nil),
			internationalCalling: util.FromOptBool(detail.Policy.Value.InternationalCalling),
			mobileSwitchToCarrier: lo.TernaryF(detail.Policy.Value.MobileSwitchToCarrier.IsSet(), func() *readDtoPolicyMobileSwitchToCarrier {
				v := detail.Policy.Value.MobileSwitchToCarrier.Value
				return &readDtoPolicyMobileSwitchToCarrier{
					enable:   util.FromOptBool(v.Enable),
					locked:   util.FromOptBool(v.Locked),
					lockedBy: util.FromOptString(v.LockedBy),
				}
			}, lo.Nil),
			selectOutboundCallerID: lo.TernaryF(detail.Policy.Value.SelectOutboundCallerID.IsSet(), func() *readDtoPolicySelectOutboundCallerID {
				v := detail.Policy.Value.SelectOutboundCallerID.Value
				return &readDtoPolicySelectOutboundCallerID{
					enable:                    util.FromOptBool(v.Enable),
					allowHideOutboundCallerID: util.FromOptBool(v.AllowHideOutboundCallerID),
					locked:                    util.FromOptBool(v.Locked),
					lockedBy:                  util.FromOptString(v.LockedBy),
				}
			}, lo.Nil),
			sms: lo.TernaryF(detail.Policy.Value.SMS.IsSet(), func() *readDtoPolicySMS {
				v := detail.Policy.Value.SMS.Value
				return &readDtoPolicySMS{
					enable:           util.FromOptBool(v.Enable),
					internationalSMS: util.FromOptBool(v.InternationalSMS),
					InternationalSMSCountries: lo.Map(v.InternationalSMSCountries, func(country string, _ int) types.String {
						return types.StringValue(country)
					}),
					locked:   util.FromOptBool(v.Locked),
					lockedBy: util.FromOptString(v.LockedBy),
				}
			}, lo.Nil),
			voicemail: lo.TernaryF(detail.Policy.Value.Voicemail.IsSet(), func() *readDtoPolicyVoicemail {
				v := detail.Policy.Value.Voicemail.Value
				return &readDtoPolicyVoicemail{
					allowDelete:        util.FromOptBool(v.AllowDelete),
					allowDownload:      util.FromOptBool(v.AllowDownload),
					allowTranscription: util.FromOptBool(v.AllowTranscription),
					allowVideomail:     util.FromOptBool(v.AllowVideomail),
					enable:             util.FromOptBool(v.Enable),
				}
			}, lo.Nil),
			voicemailAccessMembers: lo.Map(detail.Policy.Value.VoicemailAccessMembers, func(v zoomphone.PhoneUserOKPolicyVoicemailAccessMembersItem, _ int) readDtoPolicyVoicemailAccessMember {
				return readDtoPolicyVoicemailAccessMember{
					accessUserID:  util.FromOptString(v.AccessUserID),
					allowDelete:   util.FromOptBool(v.AllowDelete),
					allowDownload: util.FromOptBool(v.AllowDownload),
					allowSharing:  util.FromOptBool(v.AllowSharing),
					sharedID:      util.FromOptString(v.SharedID),
				}
			}),
			zoomPhoneOnMobile: lo.TernaryF(detail.Policy.Value.ZoomPhoneOnMobile.IsSet(), func() *readDtoPolicyZoomPhoneOnMobile {
				v := detail.Policy.Value.ZoomPhoneOnMobile.Value
				return &readDtoPolicyZoomPhoneOnMobile{
					allowCallingSMSMms: util.FromOptBool(v.AllowCallingSMSMms),
					enable:             util.FromOptBool(v.Enable),
					locked:             util.FromOptBool(v.Locked),
					lockedBy:           util.FromOptString(v.LockedBy),
				}
			}, lo.Nil),
			personalAudioLibrary: lo.TernaryF(detail.Policy.Value.PersonalAudioLibrary.IsSet(), func() *readDtoPolicyPersonalAudioLibrary {
				v := detail.Policy.Value.PersonalAudioLibrary.Value
				return &readDtoPolicyPersonalAudioLibrary{
					enable:                        util.FromOptBool(v.Enable),
					locked:                        util.FromOptBool(v.Locked),
					lockedBy:                      util.FromOptString(v.LockedBy),
					modified:                      util.FromOptBool(v.Modified),
					allowMusicOnHoldCustomization: util.FromOptBool(v.AllowMusicOnHoldCustomization),
					allowVoicemailAndMessageGreetingCustomization: util.FromOptBool(v.AllowVoicemailAndMessageGreetingCustomization),
				}
			}, lo.Nil),
			voicemailTranscription: lo.TernaryF(detail.Policy.Value.VoicemailTranscription.IsSet(), func() *readDtoPolicyVoicemailTranscription {
				v := detail.Policy.Value.VoicemailTranscription.Value
				return &readDtoPolicyVoicemailTranscription{
					enable:   util.FromOptBool(v.Enable),
					locked:   util.FromOptBool(v.Locked),
					lockedBy: util.FromOptString(v.LockedBy),
					modified: util.FromOptBool(v.Modified),
				}
			}, lo.Nil),
			voicemailNotificationByEmail: lo.TernaryF(detail.Policy.Value.VoicemailNotificationByEmail.IsSet(), func() *readDtoPolicyVoicemailNotificationByEmail {
				v := detail.Policy.Value.VoicemailNotificationByEmail.Value
				return &readDtoPolicyVoicemailNotificationByEmail{
					includeVoicemailFile:          util.FromOptBool(v.IncludeVoicemailFile),
					includeVoicemailTranscription: util.FromOptBool(v.IncludeVoicemailTranscription),
					enable:                        util.FromOptBool(v.Enable),
					locked:                        util.FromOptBool(v.Locked),
					lockedBy:                      util.FromOptString(v.LockedBy),
					modified:                      util.FromOptBool(v.Modified),
				}
			}, lo.Nil),
			sharedVoicemailNotificationByEmail: lo.TernaryF(detail.Policy.Value.SharedVoicemailNotificationByEmail.IsSet(), func() *readDtoPolicySharedVoicemailNotificationByEmail {
				v := detail.Policy.Value.SharedVoicemailNotificationByEmail.Value
				return &readDtoPolicySharedVoicemailNotificationByEmail{
					enable:   util.FromOptBool(v.Enable),
					locked:   util.FromOptBool(v.Locked),
					lockedBy: util.FromOptString(v.LockedBy),
					modified: util.FromOptBool(v.Modified),
				}
			}, lo.Nil),
			checkVoicemailsOverPhone: lo.TernaryF(detail.Policy.Value.CheckVoicemailsOverPhone.IsSet(), func() *readDtoPolicyCheckVoicemailsOverPhone {
				v := detail.Policy.Value.CheckVoicemailsOverPhone.Value
				return &readDtoPolicyCheckVoicemailsOverPhone{
					enable:   util.FromOptBool(v.Enable),
					locked:   util.FromOptBool(v.Locked),
					lockedBy: util.FromOptString(v.LockedBy),
					modified: util.FromOptBool(v.Modified),
				}
			}, lo.Nil),
			audioIntercom: lo.TernaryF(detail.Policy.Value.AudioIntercom.IsSet(), func() *readDtoPolicyAudioIntercom {
				v := detail.Policy.Value.AudioIntercom.Value
				return &readDtoPolicyAudioIntercom{
					enable:   util.FromOptBool(v.Enable),
					locked:   util.FromOptBool(v.Locked),
					lockedBy: util.FromOptString(v.LockedBy),
					modified: util.FromOptBool(v.Modified),
				}
			}, lo.Nil),
			peerToPeerMedia: lo.TernaryF(detail.Policy.Value.PeerToPeerMedia.IsSet(), func() *readDtoPolicyPeerToPeerMedia {
				v := detail.Policy.Value.PeerToPeerMedia.Value
				return &readDtoPolicyPeerToPeerMedia{
					enable:   util.FromOptBool(v.Enable),
					locked:   util.FromOptBool(v.Locked),
					lockedBy: util.FromOptString(v.LockedBy),
					modified: util.FromOptBool(v.Modified),
				}
			}, lo.Nil),
			e2eEncryption: lo.TernaryF(detail.Policy.Value.E2eEncryption.IsSet(), func() *readDtoPolicyE2eEncryption {
				v := detail.Policy.Value.E2eEncryption.Value
				return &readDtoPolicyE2eEncryption{
					enable:   util.FromOptBool(v.Enable),
					locked:   util.FromOptBool(v.Locked),
					lockedBy: util.FromOptString(v.LockedBy),
					modified: util.FromOptBool(v.Modified),
				}
			}, lo.Nil),
			outboundCalling: lo.TernaryF(detail.Policy.Value.OutboundCalling.IsSet(), func() *readDtoPolicyOutboundCalling {
				v := detail.Policy.Value.OutboundCalling.Value
				return &readDtoPolicyOutboundCalling{
					enable:   util.FromOptBool(v.Enable),
					locked:   util.FromOptBool(v.Locked),
					modified: util.FromOptBool(v.Modified),
				}
			}, lo.Nil),
			outboundSMS: lo.TernaryF(detail.Policy.Value.OutboundSMS.IsSet(), func() *readDtoPolicyOutboundSMS {
				v := detail.Policy.Value.OutboundSMS.Value
				return &readDtoPolicyOutboundSMS{
					enable:   util.FromOptBool(v.Enable),
					locked:   util.FromOptBool(v.Locked),
					modified: util.FromOptBool(v.Modified),
				}
			}, lo.Nil),
			allowEndUserEditCallHandling: lo.TernaryF(detail.Policy.Value.AllowEndUserEditCallHandling.IsSet(), func() *readDtoPolicyAllowEndUserEditCallHandling {
				v := detail.Policy.Value.AllowEndUserEditCallHandling.Value
				return &readDtoPolicyAllowEndUserEditCallHandling{
					enable:   util.FromOptBool(v.Enable),
					locked:   util.FromOptBool(v.Locked),
					modified: util.FromOptBool(v.Modified),
				}
			}, lo.Nil),
		}
	}

	return &readDto{
		callingPlans: lo.Map(detail.CallingPlans, func(callingPlan zoomphone.PhoneUserOKCallingPlansItem, _ int) readDtoCallingPlan {
			return readDtoCallingPlan{
				callingPlanType:    util.FromOptInt(callingPlan.Type),
				billingAccountID:   util.FromOptString(callingPlan.BillingAccountID),
				billingAccountName: util.FromOptString(callingPlan.BillingAccountName),
			}
		}),
		costCenter:         util.FromOptString(detail.CostCenter),
		department:         util.FromOptString(detail.Department),
		email:              util.FromOptString(detail.Email),
		emergencyAddressID: util.FromOptString(detail.EmergencyAddress.Value.ID),
		extensionID:        util.FromOptString(detail.ExtensionID),
		extensionNumber:    util.FromOptInt64(detail.ExtensionNumber),
		zoomUserID:         util.FromOptString(detail.ID),
		phoneNumbers: lo.Map(detail.PhoneNumbers, func(phoneNumber zoomphone.PhoneUserOKPhoneNumbersItem, _ int) readDtoPhoneNumber {
			return readDtoPhoneNumber{
				phoneNumberID: util.FromOptString(phoneNumber.ID),
				phoneNumber:   util.FromOptString(phoneNumber.Number),
			}
		}),
		phoneUserID: util.FromOptString(detail.PhoneUserID),
		policy:      policy,
		siteID:      util.FromOptString(detail.SiteID),
		status:      util.FromOptString(detail.Status),
	}, nil
}

func (c *userCrud) create(ctx context.Context, dto createDto) (*createdDto, error) {
	// res, err := c.phoneClient.BatchAddUsers(ctx, zoomphone.OptBatchAddUsersReq{
	// 	Value: zoomphone.BatchAddUsersReq{
	// 		Users: []zoomphone.BatchAddUsersReqUsersItem{
	// 			{
	// 				Email:           dto.email.ValueString(),
	// 				FirstName:       util.ToOptString(dto.firstName),
	// 				LastName:        util.ToOptString(dto.lastName),
	// 				CallingPlans:    []string{},
	// 				ExtensionNumber: dto.extensionNumber.ValueString(),
	// 			},
	// 		},
	// 	},
	// 	Set: true,
	// })

	// Experimental: Call Zoom User API
	err := c.userClient.UserUpdate(ctx, zoomuser.NewOptUserUpdateReq(zoomuser.UserUpdateReq{
		Feature: zoomuser.NewOptUserUpdateReqFeature(zoomuser.UserUpdateReqFeature{
			ZoomPhone: zoomuser.NewOptBool(true),
		}),
	}), zoomuser.UserUpdateParams{
		UserId: dto.zoomUserID.ValueString(),
	})

	if err != nil {
		var status *zoomuser.ErrorResponseStatusCode
		if errors.As(err, &status) {
			if status.StatusCode == 400 && status.Response.Code.Value == 404 {
				return nil, nil
			}
		}
		return nil, fmt.Errorf("error deleting phone user: %v", err)
	}

	// if err != nil {
	// 	return nil, fmt.Errorf("error creating phone user: %v", err)
	// }

	return &createdDto{
		zoomUserID: dto.zoomUserID,
	}, nil
}

func (c *userCrud) update(ctx context.Context, dto updateDto) error {
	err := c.phoneClient.UpdateUserProfile(ctx, zoomphone.NewOptUpdateUserProfileReq(
		zoomphone.UpdateUserProfileReq{
			EmergencyAddressID: util.ToOptString(dto.emergencyAddressID),
			ExtensionNumber:    util.ToOptString(dto.extensionNumber),
			Policy: lo.Ternary(dto.policy == nil,
				zoomphone.OptUpdateUserProfileReqPolicy{},
				zoomphone.NewOptUpdateUserProfileReqPolicy(
					zoomphone.UpdateUserProfileReqPolicy{
						AdHocCallRecording: lo.Ternary(dto.policy.callPark == nil,
							zoomphone.OptUpdateUserProfileReqPolicyAdHocCallRecording{},
							zoomphone.NewOptUpdateUserProfileReqPolicyAdHocCallRecording(
								zoomphone.UpdateUserProfileReqPolicyAdHocCallRecording{
									Enable:                 util.ToOptBool(dto.policy.adHocCallRecording.enable),
									RecordingStartPrompt:   util.ToOptBool(dto.policy.adHocCallRecording.recordingStartPrompt),
									RecordingTranscription: util.ToOptBool(dto.policy.adHocCallRecording.recordingTranscription),
									PlayRecordingBeepTone: lo.Ternary(dto.policy.adHocCallRecording.playRecordingBeepTone == nil,
										zoomphone.OptUpdateUserProfileReqPolicyAdHocCallRecordingPlayRecordingBeepTone{},
										zoomphone.NewOptUpdateUserProfileReqPolicyAdHocCallRecordingPlayRecordingBeepTone(
											zoomphone.UpdateUserProfileReqPolicyAdHocCallRecordingPlayRecordingBeepTone{
												Enable:               util.ToOptBool(dto.policy.adHocCallRecording.playRecordingBeepTone.enable),
												PlayBeepVolume:       util.ToOptInt(dto.policy.adHocCallRecording.playRecordingBeepTone.playBeepVolume),
												PlayBeepTimeInterval: util.ToOptInt(dto.policy.adHocCallRecording.playRecordingBeepTone.playBeepTimeInterval),
												PlayBeepMember:       util.ToOptString(dto.policy.adHocCallRecording.playRecordingBeepTone.playBeepMember),
											},
										),
									),
								},
							),
						),
						AutoCallRecording: lo.Ternary(dto.policy.autoCallRecording == nil,
							zoomphone.OptUpdateUserProfileReqPolicyAutoCallRecording{},
							zoomphone.NewOptUpdateUserProfileReqPolicyAutoCallRecording(
								zoomphone.UpdateUserProfileReqPolicyAutoCallRecording{
									AllowStopResumeRecording:     util.ToOptBool(dto.policy.autoCallRecording.allowStopResumeRecording),
									DisconnectOnRecordingFailure: util.ToOptBool(dto.policy.autoCallRecording.disconnectOnRecordingFailure),
									Enable:                       util.ToOptBool(dto.policy.autoCallRecording.enable),
									RecordingCalls:               util.ToOptString(dto.policy.autoCallRecording.recordingCalls),
									RecordingExplicitConsent:     util.ToOptBool(dto.policy.autoCallRecording.recordingExplicitConsent),
									RecordingStartPrompt:         util.ToOptBool(dto.policy.autoCallRecording.recordingStartPrompt),
									RecordingTranscription:       util.ToOptBool(dto.policy.autoCallRecording.recordingTranscription),
									PlayRecordingBeepTone: lo.Ternary(dto.policy.autoCallRecording.playRecordingBeepTone == nil,
										zoomphone.OptUpdateUserProfileReqPolicyAutoCallRecordingPlayRecordingBeepTone{},
										zoomphone.NewOptUpdateUserProfileReqPolicyAutoCallRecordingPlayRecordingBeepTone(
											zoomphone.UpdateUserProfileReqPolicyAutoCallRecordingPlayRecordingBeepTone{
												Enable:               util.ToOptBool(dto.policy.autoCallRecording.playRecordingBeepTone.enable),
												PlayBeepVolume:       util.ToOptInt(dto.policy.autoCallRecording.playRecordingBeepTone.playBeepVolume),
												PlayBeepTimeInterval: util.ToOptInt(dto.policy.autoCallRecording.playRecordingBeepTone.playBeepTimeInterval),
												PlayBeepMember:       util.ToOptString(dto.policy.autoCallRecording.playRecordingBeepTone.playBeepMember),
											},
										),
									),
								},
							),
						),
						CallOverflow: lo.Ternary(dto.policy.callOverflow == nil,
							zoomphone.OptUpdateUserProfileReqPolicyCallOverflow{},
							zoomphone.NewOptUpdateUserProfileReqPolicyCallOverflow(
								zoomphone.UpdateUserProfileReqPolicyCallOverflow{
									CallOverflowType: util.ToOptInt(dto.policy.callOverflow.callOverflowType),
									Enable:           util.ToOptBool(dto.policy.callOverflow.enable),
								},
							),
						),
						CallPark: lo.Ternary(dto.policy.callPark == nil,
							zoomphone.OptUpdateUserProfileReqPolicyCallPark{},
							zoomphone.NewOptUpdateUserProfileReqPolicyCallPark(
								zoomphone.UpdateUserProfileReqPolicyCallPark{
									CallNotPickedUpAction: util.ToOptInt(dto.policy.callPark.callNotPickedUpAction),
									Enable:                util.ToOptBool(dto.policy.callPark.enable),
									ExpirationPeriod:      util.ToOptInt(dto.policy.callPark.expirationPeriod),
									ForwardToExtensionID:  util.ToOptString(dto.policy.callPark.forwardToExtensionID),
								},
							),
						),
						CallTransferring: lo.Ternary(dto.policy.callTransferring == nil,
							zoomphone.OptUpdateUserProfileReqPolicyCallTransferring{},
							zoomphone.NewOptUpdateUserProfileReqPolicyCallTransferring(
								zoomphone.UpdateUserProfileReqPolicyCallTransferring{
									CallTransferringType: util.ToOptInt(dto.policy.callTransferring.callTransferringType),
									Enable:               util.ToOptBool(dto.policy.callTransferring.enable),
								},
							),
						),
						Delegation:       util.ToOptBool(dto.policy.delegation),
						ElevateToMeeting: util.ToOptBool(dto.policy.elevateToMeeting),
						EmergencyAddressManagement: lo.Ternary(dto.policy.emergencyAddressManagement == nil,
							zoomphone.OptUpdateUserProfileReqPolicyEmergencyAddressManagement{},
							zoomphone.NewOptUpdateUserProfileReqPolicyEmergencyAddressManagement(
								zoomphone.UpdateUserProfileReqPolicyEmergencyAddressManagement{
									Enable:               util.ToOptBool(dto.policy.emergencyAddressManagement.enable),
									PromptDefaultAddress: util.ToOptBool(dto.policy.emergencyAddressManagement.promptDefaultAddress),
								},
							),
						),
						EmergencyCallsToPsap:        util.ToOptBool(dto.policy.emergencyCallsToPsap),
						ForwardingToExternalNumbers: util.ToOptBool(dto.policy.forwardingToExternalNumbers),
						CallHandlingForwardingToOtherUsers: lo.Ternary(dto.policy.callHandlingForwardingToOtherUsers == nil,
							zoomphone.OptUpdateUserProfileReqPolicyCallHandlingForwardingToOtherUsers{},
							zoomphone.NewOptUpdateUserProfileReqPolicyCallHandlingForwardingToOtherUsers(
								zoomphone.UpdateUserProfileReqPolicyCallHandlingForwardingToOtherUsers{
									Enable:             util.ToOptBool(dto.policy.callHandlingForwardingToOtherUsers.enable),
									CallForwardingType: util.ToOptInt(dto.policy.callHandlingForwardingToOtherUsers.callForwardingType),
								},
							),
						),
						HandOffToRoom: lo.Ternary(dto.policy.handOffToRoom == nil,
							zoomphone.OptUpdateUserProfileReqPolicyHandOffToRoom{},
							zoomphone.NewOptUpdateUserProfileReqPolicyHandOffToRoom(
								zoomphone.UpdateUserProfileReqPolicyHandOffToRoom{
									Enable: util.ToOptBool(dto.policy.handOffToRoom.enable),
								},
							),
						),
						InternationalCalling: util.ToOptBool(dto.policy.internationalCalling),
						MobileSwitchToCarrier: lo.Ternary(dto.policy.mobileSwitchToCarrier == nil,
							zoomphone.OptUpdateUserProfileReqPolicyMobileSwitchToCarrier{},
							zoomphone.NewOptUpdateUserProfileReqPolicyMobileSwitchToCarrier(
								zoomphone.UpdateUserProfileReqPolicyMobileSwitchToCarrier{
									Enable: util.ToOptBool(dto.policy.mobileSwitchToCarrier.enable),
								},
							),
						),
						SelectOutboundCallerID: lo.Ternary(dto.policy.selectOutboundCallerID == nil,
							zoomphone.OptUpdateUserProfileReqPolicySelectOutboundCallerID{},
							zoomphone.NewOptUpdateUserProfileReqPolicySelectOutboundCallerID(
								zoomphone.UpdateUserProfileReqPolicySelectOutboundCallerID{
									Enable:                    util.ToOptBool(dto.policy.selectOutboundCallerID.enable),
									AllowHideOutboundCallerID: util.ToOptBool(dto.policy.selectOutboundCallerID.allowHideOutboundCallerID),
								},
							),
						),
						SMS: lo.Ternary(dto.policy.sms == nil,
							zoomphone.OptUpdateUserProfileReqPolicySMS{},
							zoomphone.NewOptUpdateUserProfileReqPolicySMS(
								zoomphone.UpdateUserProfileReqPolicySMS{
									Enable:           util.ToOptBool(dto.policy.sms.enable),
									InternationalSMS: util.ToOptBool(dto.policy.sms.internationalSMS),
									InternationalSMSCountries: lo.Map(dto.policy.sms.InternationalSMSCountries, func(v types.String, _ int) string {
										return v.ValueString()
									}),
								},
							),
						),
						Voicemail: lo.Ternary(dto.policy.voicemail == nil,
							zoomphone.OptUpdateUserProfileReqPolicyVoicemail{},
							zoomphone.NewOptUpdateUserProfileReqPolicyVoicemail(
								zoomphone.UpdateUserProfileReqPolicyVoicemail{
									AllowDelete:        util.ToOptBool(dto.policy.voicemail.allowDelete),
									AllowDownload:      util.ToOptBool(dto.policy.voicemail.allowDownload),
									AllowTranscription: util.ToOptBool(dto.policy.voicemail.allowTranscription),
									AllowVideomail:     util.ToOptBool(dto.policy.voicemail.allowVideomail),
									Enable:             util.ToOptBool(dto.policy.voicemail.enable),
								},
							),
						),
						ZoomPhoneOnMobile: lo.Ternary(dto.policy.zoomPhoneOnMobile == nil,
							zoomphone.OptUpdateUserProfileReqPolicyZoomPhoneOnMobile{},
							zoomphone.NewOptUpdateUserProfileReqPolicyZoomPhoneOnMobile(
								zoomphone.UpdateUserProfileReqPolicyZoomPhoneOnMobile{
									AllowCallingSMSMms: util.ToOptBool(dto.policy.zoomPhoneOnMobile.allowCallingSMSMms),
									Enable:             util.ToOptBool(dto.policy.zoomPhoneOnMobile.enable),
								},
							),
						),
						PersonalAudioLibrary: lo.Ternary(dto.policy.personalAudioLibrary == nil,
							zoomphone.OptUpdateUserProfileReqPolicyPersonalAudioLibrary{},
							zoomphone.NewOptUpdateUserProfileReqPolicyPersonalAudioLibrary(
								zoomphone.UpdateUserProfileReqPolicyPersonalAudioLibrary{
									AllowMusicOnHoldCustomization:                 util.ToOptBool(dto.policy.personalAudioLibrary.allowMusicOnHoldCustomization),
									AllowVoicemailAndMessageGreetingCustomization: util.ToOptBool(dto.policy.personalAudioLibrary.allowVoicemailAndMessageGreetingCustomization),
									Enable: util.ToOptBool(dto.policy.personalAudioLibrary.enable),
								},
							),
						),
						VoicemailTranscription: lo.Ternary(dto.policy.voicemailTranscription == nil,
							zoomphone.OptUpdateUserProfileReqPolicyVoicemailTranscription{},
							zoomphone.NewOptUpdateUserProfileReqPolicyVoicemailTranscription(
								zoomphone.UpdateUserProfileReqPolicyVoicemailTranscription{
									Enable: util.ToOptBool(dto.policy.voicemailTranscription.enable),
								},
							),
						),
						VoicemailNotificationByEmail: lo.Ternary(dto.policy.voicemailNotificationByEmail == nil,
							zoomphone.OptUpdateUserProfileReqPolicyVoicemailNotificationByEmail{},
							zoomphone.NewOptUpdateUserProfileReqPolicyVoicemailNotificationByEmail(
								zoomphone.UpdateUserProfileReqPolicyVoicemailNotificationByEmail{
									IncludeVoicemailFile:          util.ToOptBool(dto.policy.voicemailNotificationByEmail.includeVoicemailFile),
									IncludeVoicemailTranscription: util.ToOptBool(dto.policy.voicemailNotificationByEmail.includeVoicemailTranscription),
									Enable:                        util.ToOptBool(dto.policy.voicemailNotificationByEmail.enable),
								},
							),
						),
						SharedVoicemailNotificationByEmail: lo.Ternary(dto.policy.sharedVoicemailNotificationByEmail == nil,
							zoomphone.OptUpdateUserProfileReqPolicySharedVoicemailNotificationByEmail{},
							zoomphone.NewOptUpdateUserProfileReqPolicySharedVoicemailNotificationByEmail(
								zoomphone.UpdateUserProfileReqPolicySharedVoicemailNotificationByEmail{
									Enable: util.ToOptBool(dto.policy.sharedVoicemailNotificationByEmail.enable),
								},
							),
						),
						CheckVoicemailsOverPhone: lo.Ternary(dto.policy.checkVoicemailsOverPhone == nil,
							zoomphone.OptUpdateUserProfileReqPolicyCheckVoicemailsOverPhone{},
							zoomphone.NewOptUpdateUserProfileReqPolicyCheckVoicemailsOverPhone(
								zoomphone.UpdateUserProfileReqPolicyCheckVoicemailsOverPhone{
									Enable: util.ToOptBool(dto.policy.checkVoicemailsOverPhone.enable),
								},
							),
						),
						AudioIntercom: lo.Ternary(dto.policy.audioIntercom == nil,
							zoomphone.OptUpdateUserProfileReqPolicyAudioIntercom{},
							zoomphone.NewOptUpdateUserProfileReqPolicyAudioIntercom(
								zoomphone.UpdateUserProfileReqPolicyAudioIntercom{
									Enable: util.ToOptBool(dto.policy.audioIntercom.enable),
								},
							),
						),
						E2eEncryption: lo.Ternary(dto.policy.e2eEncryption == nil,
							zoomphone.OptUpdateUserProfileReqPolicyE2eEncryption{},
							zoomphone.NewOptUpdateUserProfileReqPolicyE2eEncryption(
								zoomphone.UpdateUserProfileReqPolicyE2eEncryption{
									Enable: util.ToOptBool(dto.policy.e2eEncryption.enable),
								},
							),
						),
					},
				),
			),
			SiteID:     util.ToOptString(dto.siteID),
			TemplateID: util.ToOptString(dto.templateID),
		},
	), zoomphone.UpdateUserProfileParams{
		UserId: dto.zoomUserID.ValueString(),
	})
	if err != nil {
		return fmt.Errorf("error updating phone user: %v", err)
	}

	return nil
}

func (c *userCrud) delete(ctx context.Context, zoomUserID string) error {
	// See also: https://devforum.zoom.us/t/remove-user-from-phone-system-management/77304/4
	err := c.userClient.UserUpdate(ctx, zoomuser.NewOptUserUpdateReq(zoomuser.UserUpdateReq{
		Feature: zoomuser.NewOptUserUpdateReqFeature(zoomuser.UserUpdateReqFeature{
			ZoomPhone: zoomuser.NewOptBool(false),
		}),
	}), zoomuser.UserUpdateParams{
		UserId: zoomUserID,
	})

	if err != nil {
		var status *zoomuser.ErrorResponseStatusCode
		if errors.As(err, &status) {
			if status.StatusCode == 400 && status.Response.Code.Value == 404 {
				return nil
			}
		}
		return fmt.Errorf("error deleting phone user: %v", err)
	}

	return nil
}
