package callhandling

import (
	"context"
	"errors"
	"fmt"

	"github.com/folio-sec/terraform-provider-zoom/generated/api/zoomphone"
	"github.com/folio-sec/terraform-provider-zoom/internal/util"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/samber/lo"
)

func newCrud(client *zoomphone.Client) *crud {
	return &crud{
		client: client,
	}
}

type crud struct {
	client *zoomphone.Client
}

// Zoom phone provides default call forwarding ids for zoom phone mobile apps, zoom phone desktop apps and zoom phone appliance client
// They become different call_forwarding_id, so collect them using description.
const callForwardingDescriptionZoomMobileApps = "Zoom Mobile Apps"
const callForwardingDescriptionZoomDesktopApps = "Zoom Desktop Apps"
const callForwardingDescriptionZoomPhoneApplianceApps = "Zoom Phone Appliance Apps"

func (c *crud) readBusinessHours(ctx context.Context, extensionID types.String) (*readDtoBusinessHours, error) {
	detail, err := c.client.GetCallHandling(ctx, zoomphone.GetCallHandlingParams{
		ExtensionId: extensionID.ValueString(),
	})
	if err != nil {
		var status *zoomphone.ErrorResponseStatusCode
		if errors.As(err, &status) {
			if status.StatusCode == 400 && status.Response.Code.Value == 300 {
				return nil, nil // already deleted
			}
		}
		return nil, fmt.Errorf("unable to read phone call handling: %v", err)
	}

	// BusinessHours should contain upper to one custom_hours
	retCustomHour, _ := lo.Find(detail.BusinessHours, func(item zoomphone.GetCallHandlingOKBusinessHoursItem) bool {
		return item.SubSettingType.Value == "custom_hours"
	})
	customHours := &readDtoBusinessHoursCustomHours{
		typ:                 util.FromOptInt(retCustomHour.Settings.Value.Type),
		allowMembersToReset: util.FromOptBool(retCustomHour.Settings.Value.AllowMembersToReset),
		settings: lo.Map(retCustomHour.Settings.Value.CustomHoursSettings, func(item zoomphone.GetCallHandlingOKBusinessHoursItemSettingsCustomHoursSettingsItem, index int) *readDtoBusinessHoursCustomHoursSetting {
			return &readDtoBusinessHoursCustomHoursSetting{
				weekday: util.FromOptInt(item.Weekday),
				typ:     util.FromOptInt(item.Type),
				from:    util.FromOptString(item.From),
				to:      util.FromOptString(item.To),
			}
		}),
	}

	// BusinessHours should contain upper to one call_handling.
	retCallHandling, _ := lo.Find(detail.BusinessHours, func(item zoomphone.GetCallHandlingOKBusinessHoursItem) bool {
		return item.SubSettingType.Value == "call_handling"
	})
	forwardToExtensionID := types.StringNull()
	if len(retCallHandling.Settings.Value.CallForwardingSettings) > 0 {
		forwardToExtensionID = util.FromOptString(retCallHandling.Settings.Value.CallForwardingSettings[0].ID)
	}
	callDistribution := &readDtoBusinessHoursCallHandlingCallDistribution{
		handleMultipleCalls:          util.FromOptBool(retCallHandling.Settings.Value.CallDistribution.Value.HandleMultipleCalls),
		ringDuration:                 util.FromOptInt(retCallHandling.Settings.Value.CallDistribution.Value.RingDuration),
		ringMode:                     util.FromOptString(retCallHandling.Settings.Value.CallDistribution.Value.RingMode),
		skipOfflineDevicePhoneNumber: util.FromOptBool(retCallHandling.Settings.Value.CallDistribution.Value.SkipOfflineDevicePhoneNumber),
	}
	tflog.Info(ctx, "readBusinessHours", map[string]interface{}{
		"retCallHandling": retCallHandling,
		"settings":        retCallHandling.Settings,
		"receiveCall":     retCallHandling.Settings.Value.ReceiveCall.Value,
	})
	callHandling := &readDtoBusinessHoursCallHandling{
		callNotAnswerAction:                     util.FromOptInt(retCallHandling.Settings.Value.CallNotAnswerAction),
		forwardToExtensionID:                    forwardToExtensionID,
		busyOnAnotherCallAction:                 util.FromOptInt(retCallHandling.Settings.Value.BusyRouting.Value.Action),
		busyForwardToExtensionID:                util.FromOptString(retCallHandling.Settings.Value.BusyRouting.Value.ForwardTo.Value.ExtensionID),
		allowCallersCheckVoicemail:              util.FromOptBool(retCallHandling.Settings.Value.AllowCallersCheckVoicemail),
		allowMembersToReset:                     util.FromOptBool(retCallHandling.Settings.Value.AllowMembersToReset),
		audioWhileConnectingID:                  util.FromOptString(retCallHandling.Settings.Value.AudioWhileConnecting.Value.ID),
		callDistribution:                        callDistribution,
		busyRequirePress1BeforeConnecting:       util.FromOptBool(retCallHandling.Settings.Value.BusyRouting.Value.RequirePress1BeforeConnecting),
		unAnsweredRequirePress1BeforeConnecting: util.FromOptBool(retCallHandling.Settings.Value.Routing.Value.RequirePress1BeforeConnecting),
		overflowPlayCalleeVoicemailGreeting:     util.FromOptBool(retCallHandling.Settings.Value.Routing.Value.OverflowPlayCalleeVoicemailGreeting),
		playCalleeVoicemailGreeting:             util.FromOptBool(retCallHandling.Settings.Value.Routing.Value.OverflowPlayCalleeVoicemailGreeting),
		busyPlayCalleeVoicemailGreeting:         util.FromOptBool(retCallHandling.Settings.Value.BusyRouting.Value.PlayCalleeVoicemailGreeting),
		phoneNumber:                             util.FromOptString(retCallHandling.Settings.Value.Routing.Value.ForwardTo.Value.PhoneNumber),
		phoneNumberDescription:                  util.FromOptString(retCallHandling.Settings.Value.Routing.Value.ForwardTo.Value.Description),
		busyPhoneNumber:                         util.FromOptString(retCallHandling.Settings.Value.BusyRouting.Value.ForwardTo.Value.PhoneNumber),
		busyPhoneNumberDescription:              util.FromOptString(retCallHandling.Settings.Value.BusyRouting.Value.ForwardTo.Value.Description),
		connectToOperator:                       util.FromOptBool(retCallHandling.Settings.Value.ConnectToOperator),
		greetingPromptID:                        util.FromOptString(retCallHandling.Settings.Value.GreetingPrompt.Value.ID),
		maxCallInQueue:                          util.FromOptInt(retCallHandling.Settings.Value.MaxCallInQueue),
		maxWaitTime:                             util.FromOptInt(retCallHandling.Settings.Value.MaxWaitTime),
		musicOnHoldID:                           util.FromOptString(retCallHandling.Settings.Value.MusicOnHold.Value.ID),
		operatorExtensionID:                     util.FromOptString(retCallHandling.Settings.Value.Routing.Value.Operator.Value.ExtensionID),
		receiveCall:                             util.FromOptBool(retCallHandling.Settings.Value.ReceiveCall),
		ringMode:                                util.FromOptString(retCallHandling.Settings.Value.RingMode),
		voiceMailGreetingID:                     util.FromOptString(retCallHandling.Settings.Value.GreetingPrompt.Value.ID),
		wrapUpTime:                              util.FromOptInt(retCallHandling.Settings.Value.WrapUpTime),
	}

	// BusinessHours should contain upper to one call_forwarding.
	var callForwarding *readDtoHolidayHoursCallForwarding
	retCallForwarding, ok := lo.Find(detail.BusinessHours, func(item zoomphone.GetCallHandlingOKBusinessHoursItem) bool {
		return item.SubSettingType.Value == "call_forwarding"
	})
	if ok {
		enableZoomMobileApps, enableZoomDesktopApps, enableZoomPhoneApplianceApps := types.BoolNull(), types.BoolNull(), types.BoolNull()
		zoomPhoneMobileAppsItem, okMobile := lo.Find(retCallForwarding.Settings.Value.CallForwardingSettings, func(item zoomphone.GetCallHandlingOKBusinessHoursItemSettingsCallForwardingSettingsItem) bool {
			return item.Description.Value == callForwardingDescriptionZoomMobileApps
		})
		if okMobile {
			enableZoomMobileApps = util.FromOptBool(zoomPhoneMobileAppsItem.Enable)
		}
		zoomPhoneDesktopAppsItem, okDesktop := lo.Find(retCallForwarding.Settings.Value.CallForwardingSettings, func(item zoomphone.GetCallHandlingOKBusinessHoursItemSettingsCallForwardingSettingsItem) bool {
			return item.Description.Value == callForwardingDescriptionZoomDesktopApps
		})
		if okDesktop {
			enableZoomDesktopApps = util.FromOptBool(zoomPhoneDesktopAppsItem.Enable)
		}
		zoomPhoneApplianceAppsItem, okAppliance := lo.Find(retCallForwarding.Settings.Value.CallForwardingSettings, func(item zoomphone.GetCallHandlingOKBusinessHoursItemSettingsCallForwardingSettingsItem) bool {
			return item.Description.Value == callForwardingDescriptionZoomPhoneApplianceApps
		})
		if okAppliance {
			enableZoomPhoneApplianceApps = util.FromOptBool(zoomPhoneApplianceAppsItem.Enable)
		}
		callForwarding = &readDtoHolidayHoursCallForwarding{
			requirePress1BeforeConnecting: util.FromOptBool(retCallForwarding.Settings.Value.RequirePress1BeforeConnecting),
			enableZoomMobileApps:          enableZoomMobileApps,
			enableZoomDesktopApps:         enableZoomDesktopApps,
			enableZoomPhoneApplianceApps:  enableZoomPhoneApplianceApps,
			settings: lo.FilterMap(retCallForwarding.Settings.Value.CallForwardingSettings, func(item zoomphone.GetCallHandlingOKBusinessHoursItemSettingsCallForwardingSettingsItem, index int) (*readDtoCallForwardingSetting, bool) {
				description := item.Description.Value
				// Zoom provided call_forwarding setting should be ignored (they are managed by enableZoomMobileApps and so on)
				isManaged := (description != callForwardingDescriptionZoomMobileApps &&
					description != callForwardingDescriptionZoomDesktopApps &&
					description != callForwardingDescriptionZoomPhoneApplianceApps) && item.PhoneNumber.IsSet()
				return &readDtoCallForwardingSetting{
					id:          util.FromOptString(item.ID),
					description: util.FromOptString(item.Description),
					enable:      util.FromOptBool(item.Enable),
					phoneNumber: util.FromOptString(item.PhoneNumber),
					externalContact: &readDtoCallForwardingSettingsExternalContact{
						externalContactID: util.FromOptString(item.ExternalContact.Value.ExternalContactID),
					},
				}, isManaged
			}),
		}
	}

	return &readDtoBusinessHours{
		extensionID:    extensionID,
		customHours:    customHours,
		callHandling:   callHandling,
		callForwarding: callForwarding,
	}, nil
}

func (c *crud) readClosedHours(ctx context.Context, extensionID types.String) (*readDtoClosedHours, error) {
	detail, err := c.client.GetCallHandling(ctx, zoomphone.GetCallHandlingParams{
		ExtensionId: extensionID.ValueString(),
	})
	if err != nil {
		var status *zoomphone.ErrorResponseStatusCode
		if errors.As(err, &status) {
			if status.StatusCode == 400 && status.Response.Code.Value == 300 {
				return nil, nil // already deleted
			}
		}
		return nil, fmt.Errorf("unable to read phone call handling: %v", err)
	}

	// ClosedHours should contain upper to one call_handling.
	retCallHandling, _ := lo.Find(detail.ClosedHours, func(item zoomphone.GetCallHandlingOKClosedHoursItem) bool {
		return item.SubSettingType.Value == "call_handling"
	})
	forwardToExtensionID := types.StringNull()
	if len(retCallHandling.Settings.Value.CallForwardingSettings) > 0 {
		forwardToExtensionID = util.FromOptString(retCallHandling.Settings.Value.CallForwardingSettings[0].ID)
	}
	callHandling := &readDtoClosedHoursCallHandling{
		callNotAnswerAction:                     util.FromOptInt(retCallHandling.Settings.Value.CallNotAnswerAction),
		forwardToExtensionID:                    forwardToExtensionID,
		busyOnAnotherCallAction:                 util.FromOptInt(retCallHandling.Settings.Value.BusyRouting.Value.Action),
		busyForwardToExtensionID:                util.FromOptString(retCallHandling.Settings.Value.BusyRouting.Value.ForwardTo.Value.ExtensionID),
		allowCallersCheckVoicemail:              util.FromOptBool(retCallHandling.Settings.Value.AllowCallersCheckVoicemail),
		busyRequirePress1BeforeConnecting:       util.FromOptBool(retCallHandling.Settings.Value.BusyRouting.Value.RequirePress1BeforeConnecting),
		unAnsweredRequirePress1BeforeConnecting: util.FromOptBool(retCallHandling.Settings.Value.Routing.Value.RequirePress1BeforeConnecting),
		overflowPlayCalleeVoicemailGreeting:     util.FromOptBool(retCallHandling.Settings.Value.Routing.Value.OverflowPlayCalleeVoicemailGreeting),
		playCalleeVoicemailGreeting:             util.FromOptBool(retCallHandling.Settings.Value.Routing.Value.OverflowPlayCalleeVoicemailGreeting),
		busyPlayCalleeVoicemailGreeting:         util.FromOptBool(retCallHandling.Settings.Value.BusyRouting.Value.PlayCalleeVoicemailGreeting),
		phoneNumber:                             util.FromOptString(retCallHandling.Settings.Value.Routing.Value.ForwardTo.Value.PhoneNumber),
		phoneNumberDescription:                  util.FromOptString(retCallHandling.Settings.Value.Routing.Value.ForwardTo.Value.Description),
		busyPhoneNumber:                         util.FromOptString(retCallHandling.Settings.Value.BusyRouting.Value.ForwardTo.Value.PhoneNumber),
		busyPhoneNumberDescription:              util.FromOptString(retCallHandling.Settings.Value.BusyRouting.Value.ForwardTo.Value.Description),
		connectToOperator:                       util.FromOptBool(retCallHandling.Settings.Value.ConnectToOperator),
		maxWaitTime:                             util.FromOptInt(retCallHandling.Settings.Value.MaxWaitTime),
		operatorExtensionID:                     util.FromOptString(retCallHandling.Settings.Value.Routing.Value.Operator.Value.ExtensionID),
		ringMode:                                util.FromOptString(retCallHandling.Settings.Value.RingMode),
	}

	// ClosedHours should contain upper to one call_forwarding.
	var callForwarding *readDtoHolidayHoursCallForwarding
	retCallForwarding, ok := lo.Find(detail.ClosedHours, func(item zoomphone.GetCallHandlingOKClosedHoursItem) bool {
		return item.SubSettingType.Value == "call_forwarding"
	})
	if ok {
		enableZoomMobileApps, enableZoomDesktopApps, enableZoomPhoneApplianceApps := types.BoolNull(), types.BoolNull(), types.BoolNull()
		zoomPhoneMobileAppsItem, okMobile := lo.Find(retCallForwarding.Settings.Value.CallForwardingSettings, func(item zoomphone.GetCallHandlingOKClosedHoursItemSettingsCallForwardingSettingsItem) bool {
			return item.Description.Value == callForwardingDescriptionZoomMobileApps
		})
		if okMobile {
			enableZoomMobileApps = util.FromOptBool(zoomPhoneMobileAppsItem.Enable)
		}
		zoomPhoneDesktopAppsItem, okDesktop := lo.Find(retCallForwarding.Settings.Value.CallForwardingSettings, func(item zoomphone.GetCallHandlingOKClosedHoursItemSettingsCallForwardingSettingsItem) bool {
			return item.Description.Value == callForwardingDescriptionZoomDesktopApps
		})
		if okDesktop {
			enableZoomDesktopApps = util.FromOptBool(zoomPhoneDesktopAppsItem.Enable)
		}
		zoomPhoneApplianceAppsItem, okAppliance := lo.Find(retCallForwarding.Settings.Value.CallForwardingSettings, func(item zoomphone.GetCallHandlingOKClosedHoursItemSettingsCallForwardingSettingsItem) bool {
			return item.Description.Value == callForwardingDescriptionZoomPhoneApplianceApps
		})
		if okAppliance {
			enableZoomPhoneApplianceApps = util.FromOptBool(zoomPhoneApplianceAppsItem.Enable)
		}
		callForwarding = &readDtoHolidayHoursCallForwarding{
			requirePress1BeforeConnecting: util.FromOptBool(retCallForwarding.Settings.Value.RequirePress1BeforeConnecting),
			enableZoomMobileApps:          enableZoomMobileApps,
			enableZoomDesktopApps:         enableZoomDesktopApps,
			enableZoomPhoneApplianceApps:  enableZoomPhoneApplianceApps,
			settings: lo.FilterMap(retCallForwarding.Settings.Value.CallForwardingSettings, func(item zoomphone.GetCallHandlingOKClosedHoursItemSettingsCallForwardingSettingsItem, index int) (*readDtoCallForwardingSetting, bool) {
				description := item.Description.Value
				// Zoom provided call_forwarding setting should be ignored (they are managed by enableZoomMobileApps and so on)
				isManaged := (description != callForwardingDescriptionZoomMobileApps &&
					description != callForwardingDescriptionZoomDesktopApps &&
					description != callForwardingDescriptionZoomPhoneApplianceApps) && item.PhoneNumber.IsSet()
				return &readDtoCallForwardingSetting{
					id:          util.FromOptString(item.ID),
					description: util.FromOptString(item.Description),
					enable:      util.FromOptBool(item.Enable),
					phoneNumber: util.FromOptString(item.PhoneNumber),
					externalContact: &readDtoCallForwardingSettingsExternalContact{
						externalContactID: util.FromOptString(item.ExternalContact.Value.ExternalContactID),
					},
				}, isManaged
			}),
		}
	}
	return &readDtoClosedHours{
		extensionID:    extensionID,
		callHandling:   callHandling,
		callForwarding: callForwarding,
	}, nil
}

func (c *crud) readHolidayHours(ctx context.Context, extensionID, holidayID types.String) (*readDtoHolidayHours, error) {
	detail, err := c.client.GetCallHandling(ctx, zoomphone.GetCallHandlingParams{
		ExtensionId: extensionID.ValueString(),
	})
	if err != nil {
		var status *zoomphone.ErrorResponseStatusCode
		if errors.As(err, &status) {
			if status.StatusCode == 400 && status.Response.Code.Value == 300 {
				return nil, nil // already deleted
			}
		}
		return nil, fmt.Errorf("unable to read phone call handling: %v", err)
	}

	// holiday may contain multiple settings, so filtered by holiday id
	target, ok := lo.Find(detail.HolidayHours, func(item zoomphone.GetCallHandlingOKHolidayHoursItem) bool {
		return item.HolidayID.Value == holidayID.ValueString()
	})
	if !ok {
		return nil, nil // not existed
	}

	// target should contain upper to one holiday setting
	retHoliday, _ := lo.Find(target.Details, func(item zoomphone.GetCallHandlingOKHolidayHoursItemDetailsItem) bool {
		return item.SubSettingType.Value == "holiday"
	})
	holiday := &readDtoHolidayHoursHoliday{
		name: util.FromOptString(retHoliday.Settings.Value.Name),
		from: util.FromOptDateTime(retHoliday.Settings.Value.From),
		to:   util.FromOptDateTime(retHoliday.Settings.Value.To),
	}

	// target should contain upper to one call_handling setting
	retCallHandling, _ := lo.Find(target.Details, func(item zoomphone.GetCallHandlingOKHolidayHoursItemDetailsItem) bool {
		return item.SubSettingType.Value == "call_handling"
	})
	forwardToExtensionID := types.StringNull()
	if len(retCallHandling.Settings.Value.CallForwardingSettings) > 0 {
		forwardToExtensionID = util.FromOptString(retCallHandling.Settings.Value.CallForwardingSettings[0].ID)
	}
	callHandling := &readDtoHolidayHoursCallHandling{
		callNotAnswerAction:                     util.FromOptInt(retCallHandling.Settings.Value.CallNotAnswerAction),
		forwardToExtensionID:                    forwardToExtensionID,
		allowCallersCheckVoicemail:              util.FromOptBool(retCallHandling.Settings.Value.AllowCallersCheckVoicemail),
		unAnsweredRequirePress1BeforeConnecting: util.FromOptBool(retCallHandling.Settings.Value.Routing.Value.RequirePress1BeforeConnecting),
		overflowPlayCalleeVoicemailGreeting:     util.FromOptBool(retCallHandling.Settings.Value.Routing.Value.OverflowPlayCalleeVoicemailGreeting),
		playCalleeVoicemailGreeting:             util.FromOptBool(retCallHandling.Settings.Value.Routing.Value.OverflowPlayCalleeVoicemailGreeting),
		phoneNumber:                             util.FromOptString(retCallHandling.Settings.Value.Routing.Value.ForwardTo.Value.PhoneNumber),
		phoneNumberDescription:                  util.FromOptString(retCallHandling.Settings.Value.Routing.Value.ForwardTo.Value.Description),
		connectToOperator:                       util.FromOptBool(retCallHandling.Settings.Value.ConnectToOperator),
		maxWaitTime:                             util.FromOptInt(retCallHandling.Settings.Value.MaxWaitTime),
		operatorExtensionID:                     util.FromOptString(retCallHandling.Settings.Value.Routing.Value.Operator.Value.ExtensionID),
		ringMode:                                util.FromOptString(retCallHandling.Settings.Value.RingMode),
	}

	// target should contain upper to one call_forwarding setting
	var callForwarding *readDtoHolidayHoursCallForwarding
	retCallForwarding, ok := lo.Find(target.Details, func(item zoomphone.GetCallHandlingOKHolidayHoursItemDetailsItem) bool {
		return item.SubSettingType.Value == "call_forwarding"
	})
	if ok {
		enableZoomMobileApps, enableZoomDesktopApps, enableZoomPhoneApplianceApps := types.BoolNull(), types.BoolNull(), types.BoolNull()
		zoomPhoneMobileAppsItem, okMobile := lo.Find(retCallForwarding.Settings.Value.CallForwardingSettings, func(item zoomphone.GetCallHandlingOKHolidayHoursItemDetailsItemSettingsCallForwardingSettingsItem) bool {
			return item.Description.Value == callForwardingDescriptionZoomMobileApps
		})
		if okMobile {
			enableZoomMobileApps = util.FromOptBool(zoomPhoneMobileAppsItem.Enable)
		}
		zoomPhoneDesktopAppsItem, okDesktop := lo.Find(retCallForwarding.Settings.Value.CallForwardingSettings, func(item zoomphone.GetCallHandlingOKHolidayHoursItemDetailsItemSettingsCallForwardingSettingsItem) bool {
			return item.Description.Value == callForwardingDescriptionZoomDesktopApps
		})
		if okDesktop {
			enableZoomDesktopApps = util.FromOptBool(zoomPhoneDesktopAppsItem.Enable)
		}
		zoomPhoneApplianceAppsItem, okAppliance := lo.Find(retCallForwarding.Settings.Value.CallForwardingSettings, func(item zoomphone.GetCallHandlingOKHolidayHoursItemDetailsItemSettingsCallForwardingSettingsItem) bool {
			return item.Description.Value == callForwardingDescriptionZoomPhoneApplianceApps
		})
		if okAppliance {
			enableZoomPhoneApplianceApps = util.FromOptBool(zoomPhoneApplianceAppsItem.Enable)
		}
		callForwarding = &readDtoHolidayHoursCallForwarding{
			requirePress1BeforeConnecting: util.FromOptBool(retCallForwarding.Settings.Value.RequirePress1BeforeConnecting),
			enableZoomMobileApps:          enableZoomMobileApps,
			enableZoomDesktopApps:         enableZoomDesktopApps,
			enableZoomPhoneApplianceApps:  enableZoomPhoneApplianceApps,
			settings: lo.FilterMap(retCallForwarding.Settings.Value.CallForwardingSettings, func(item zoomphone.GetCallHandlingOKHolidayHoursItemDetailsItemSettingsCallForwardingSettingsItem, index int) (*readDtoCallForwardingSetting, bool) {
				description := item.Description.Value
				// Zoom provided call_forwarding setting should be ignored (they are managed by enableZoomMobileApps and so on)
				isManaged := (description != callForwardingDescriptionZoomMobileApps &&
					description != callForwardingDescriptionZoomDesktopApps &&
					description != callForwardingDescriptionZoomPhoneApplianceApps) && item.PhoneNumber.IsSet()
				return &readDtoCallForwardingSetting{
					id:          util.FromOptString(item.ID),
					description: util.FromOptString(item.Description),
					enable:      util.FromOptBool(item.Enable),
					phoneNumber: util.FromOptString(item.PhoneNumber),
					externalContact: &readDtoCallForwardingSettingsExternalContact{
						externalContactID: util.FromOptString(item.ExternalContact.Value.ExternalContactID),
					},
				}, isManaged
			}),
		}
	}

	return &readDtoHolidayHours{
		extensionID:    extensionID,
		holiday:        holiday,
		callHandling:   callHandling,
		callForwarding: callForwarding,
	}, nil
}

func (c *crud) createHoliday(ctx context.Context, dto *createHolidayDto) (*createdHolidayDto, error) {
	res, err := c.client.AddCallHandling(ctx, zoomphone.OptAddCallHandlingReq{
		Value: zoomphone.AddCallHandlingReq{
			Type: zoomphone.PostCallHandlingSettingsHolidayAddCallHandlingReq,
			PostCallHandlingSettingsHoliday: zoomphone.PostCallHandlingSettingsHoliday{
				Settings: zoomphone.OptPostCallHandlingSettingsHolidaySettings{
					Value: zoomphone.PostCallHandlingSettingsHolidaySettings{
						Name: util.ToPhoneOptString(dto.name),
						From: util.ToPhoneOptDateTime(dto.from),
						To:   util.ToPhoneOptDateTime(dto.to),
					},
					Set: true,
				},
				SubSettingType: zoomphone.NewOptString("holiday"),
			},
		},
		Set: true,
	}, zoomphone.AddCallHandlingParams{
		ExtensionId: dto.extensionID.ValueString(),
		SettingType: string(dto.settingType),
	})
	if err != nil {
		return nil, fmt.Errorf("error creating phone call handling on holiday: %v", err)
	}

	return &createdHolidayDto{
		holidayID: util.FromOptString(res.AddCallHandlingCreated1.HolidayID),
	}, nil
}

func (c *crud) createCallForwarding(ctx context.Context, dto *createCallForwardingDto) (*createdCallForwardingDto, error) {
	res, err := c.client.AddCallHandling(ctx, zoomphone.OptAddCallHandlingReq{
		Value: zoomphone.AddCallHandlingReq{
			Type: zoomphone.PostCallHandlingSettingsCallForwardingAddCallHandlingReq,
			PostCallHandlingSettingsCallForwarding: zoomphone.PostCallHandlingSettingsCallForwarding{
				Settings: zoomphone.OptPostCallHandlingSettingsCallForwardingSettings{
					Value: zoomphone.PostCallHandlingSettingsCallForwardingSettings{
						HolidayID:   util.ToPhoneOptString(dto.holidayID),
						Description: util.ToPhoneOptString(dto.description),
						PhoneNumber: util.ToPhoneOptString(dto.phoneNumber),
					},
					Set: true,
				},
				SubSettingType: zoomphone.NewOptString("call_forwarding"),
			},
		},
		Set: true,
	}, zoomphone.AddCallHandlingParams{
		ExtensionId: dto.extensionID.ValueString(),
		SettingType: string(dto.settingType),
	})
	if err != nil {
		return nil, fmt.Errorf("error creating phone call handling on call forwarding: %v", err)
	}

	return &createdCallForwardingDto{
		callForwardingID: util.FromOptString(res.AddCallHandlingCreated0.CallForwardingID),
	}, nil
}

func (c *crud) patchCustomHours(ctx context.Context, dto *patchCustomHoursDto) error {
	err := c.client.UpdateCallHandling(ctx, zoomphone.OptUpdateCallHandlingReq{
		Value: zoomphone.UpdateCallHandlingReq{
			Type: zoomphone.PatchCallHandlingSettingsCustomHoursUpdateCallHandlingReq,
			PatchCallHandlingSettingsCustomHours: zoomphone.PatchCallHandlingSettingsCustomHours{
				Settings: zoomphone.OptPatchCallHandlingSettingsCustomHoursSettings{
					Value: zoomphone.PatchCallHandlingSettingsCustomHoursSettings{
						AllowMembersToReset: util.ToPhoneOptBool(dto.allowMembersToReset),
						CustomHoursSettings: lo.Map(dto.settings, func(item *patchCustomHoursDtoSetting, index int) zoomphone.PatchCallHandlingSettingsCustomHoursSettingsCustomHoursSettingsItem {
							return zoomphone.PatchCallHandlingSettingsCustomHoursSettingsCustomHoursSettingsItem{
								From:    util.ToPhoneOptString(item.from),
								To:      util.ToPhoneOptString(item.to),
								Type:    util.ToPhoneOptInt(item.typ),
								Weekday: util.ToPhoneOptInt(item.weekday),
							}
						}),
						Type: util.ToPhoneOptInt(dto.typ),
					},
					Set: true,
				},
				SubSettingType: zoomphone.NewOptString("custom_hours"),
			},
		},
		Set: true,
	}, zoomphone.UpdateCallHandlingParams{
		ExtensionId: dto.extensionID.ValueString(),
		SettingType: string(dto.settingType),
	})
	if err != nil {
		return fmt.Errorf("error patching phone call handling on custom hour: %v", err)
	}

	return nil
}

func (c *crud) patchCallHandling(ctx context.Context, dto *patchCallHandlingDto) error {
	err := c.client.UpdateCallHandling(ctx, zoomphone.OptUpdateCallHandlingReq{
		Value: zoomphone.UpdateCallHandlingReq{
			Type: zoomphone.PatchCallHandlingSettingsCallHandlingUpdateCallHandlingReq,
			PatchCallHandlingSettingsCallHandling: zoomphone.PatchCallHandlingSettingsCallHandling{
				Settings: zoomphone.OptPatchCallHandlingSettingsCallHandlingSettings{
					Value: zoomphone.PatchCallHandlingSettingsCallHandlingSettings{
						HolidayID:                  util.ToPhoneOptString(dto.settings.holidayID),
						AllowCallersCheckVoicemail: util.ToPhoneOptBool(dto.settings.allowCallersCheckVoicemail),
						AllowMembersToReset:        util.ToPhoneOptBool(dto.settings.allowMembersToReset),
						AudioWhileConnectingID:     util.ToPhoneOptString(dto.settings.audioWhileConnectingID),
						CallDistribution: lo.TernaryF(dto.settings.callDistribution != nil, func() zoomphone.OptPatchCallHandlingSettingsCallHandlingSettingsCallDistribution {
							return zoomphone.OptPatchCallHandlingSettingsCallHandlingSettingsCallDistribution{
								Value: zoomphone.PatchCallHandlingSettingsCallHandlingSettingsCallDistribution{
									HandleMultipleCalls:          util.ToPhoneOptBool(dto.settings.callDistribution.handleMultipleCalls),
									RingDuration:                 util.ToPhoneOptInt(dto.settings.callDistribution.ringDuration),
									RingMode:                     util.ToPhoneOptString(dto.settings.callDistribution.ringMode),
									SkipOfflineDevicePhoneNumber: util.ToPhoneOptBool(dto.settings.callDistribution.skipOfflineDevicePhoneNumber),
								},
								Set: true,
							}
						}, func() zoomphone.OptPatchCallHandlingSettingsCallHandlingSettingsCallDistribution {
							return zoomphone.OptPatchCallHandlingSettingsCallHandlingSettingsCallDistribution{}
						}),
						CallNotAnswerAction:                     util.ToPhoneOptInt(dto.settings.callNotAnswerAction),
						BusyOnAnotherCallAction:                 util.ToPhoneOptInt(dto.settings.busyOnAnotherCallAction),
						BusyRequirePress1BeforeConnecting:       util.ToPhoneOptBool(dto.settings.busyRequirePress1BeforeConnecting),
						UnAnsweredRequirePress1BeforeConnecting: util.ToPhoneOptBool(dto.settings.unAnsweredRequirePress1BeforeConnecting),
						OverflowPlayCalleeVoicemailGreeting:     util.ToPhoneOptBool(dto.settings.overflowPlayCalleeVoicemailGreeting),
						PlayCalleeVoicemailGreeting:             util.ToPhoneOptBool(dto.settings.playCalleeVoicemailGreeting),
						BusyPlayCalleeVoicemailGreeting:         util.ToPhoneOptBool(dto.settings.busyPlayCalleeVoicemailGreeting),
						PhoneNumber:                             util.ToPhoneOptString(dto.settings.phoneNumber),
						Description:                             util.ToPhoneOptString(dto.settings.phoneNumberDescription),
						BusyPhoneNumber:                         util.ToPhoneOptString(dto.settings.busyPhoneNumber),
						BusyDescription:                         util.ToPhoneOptString(dto.settings.busyPhoneNumberDescription),
						ConnectToOperator:                       util.ToPhoneOptBool(dto.settings.connectToOperator),
						ForwardToExtensionID:                    util.ToPhoneOptString(dto.settings.forwardToExtensionID),
						BusyForwardToExtensionID:                util.ToPhoneOptString(dto.settings.busyForwardToExtensionID),
						GreetingPromptID:                        util.ToPhoneOptString(dto.settings.greetingPromptID),
						MaxCallInQueue:                          util.ToPhoneOptInt(dto.settings.maxCallInQueue),
						MaxWaitTime:                             util.ToPhoneOptInt(dto.settings.maxWaitTime),
						MusicOnHoldID:                           util.ToPhoneOptString(dto.settings.musicOnHoldID),
						OperatorExtensionID:                     util.ToPhoneOptString(dto.settings.operatorExtensionID),
						ReceiveCall:                             util.ToPhoneOptBool(dto.settings.receiveCall),
						RingMode:                                util.ToPhoneOptString(dto.settings.ringMode),
						VoicemailGreetingID:                     util.ToPhoneOptString(dto.settings.voiceMailGreetingID),
						WrapUpTime:                              util.ToPhoneOptInt(dto.settings.wrapUpTime),
					},
					Set: true,
				},
				SubSettingType: zoomphone.NewOptString("call_handling"),
			},
		},
		Set: true,
	}, zoomphone.UpdateCallHandlingParams{
		ExtensionId: dto.extensionID.ValueString(),
		SettingType: string(dto.settingType),
	})
	if err != nil {
		return fmt.Errorf("error patching phone call handling on call handling: %v", err)
	}

	return nil
}

func (c *crud) patchHoliday(ctx context.Context, dto *patchHolidayDto) error {
	err := c.client.UpdateCallHandling(ctx, zoomphone.OptUpdateCallHandlingReq{
		Value: zoomphone.UpdateCallHandlingReq{
			Type: zoomphone.PatchCallHandlingSettingsHolidayUpdateCallHandlingReq,
			PatchCallHandlingSettingsHoliday: zoomphone.PatchCallHandlingSettingsHoliday{
				Settings: zoomphone.OptPatchCallHandlingSettingsHolidaySettings{
					Value: zoomphone.PatchCallHandlingSettingsHolidaySettings{
						HolidayID: util.ToPhoneOptString(dto.holidayID),
						Name:      util.ToPhoneOptString(dto.name),
						From:      util.ToPhoneOptDateTime(dto.from),
						To:        util.ToPhoneOptDateTime(dto.to),
					},
					Set: true,
				},
				SubSettingType: zoomphone.NewOptString("holiday"),
			},
		},
		Set: true,
	}, zoomphone.UpdateCallHandlingParams{
		ExtensionId: dto.extensionID.ValueString(),
		SettingType: string(dto.settingType),
	})
	if err != nil {
		return fmt.Errorf("error patching phone call handling custom hour: %v", err)
	}

	return nil
}

func (c *crud) patchCallForwarding(ctx context.Context, dto *patchCallForwardingDto, onDelete bool) error {
	// to patch call forwarding, collect zoom predefined call forwarding settings
	// such as Zoom Mobile Apps, Zoom Desktop Apps, Zoom Phone Appliance Apps
	detail, err := c.client.GetCallHandling(ctx, zoomphone.GetCallHandlingParams{
		ExtensionId: dto.extensionID.ValueString(),
	})
	if err != nil {
		return err
	}
	var callForwardingSettings []zoomphone.PatchCallHandlingSettingsCallForwardingSettingsCallForwardingSettingsItem
	switch dto.settingType {
	case settingTypeBusinessHours:
		callForwarding, ok := lo.Find(detail.BusinessHours, func(item zoomphone.GetCallHandlingOKBusinessHoursItem) bool {
			return item.SubSettingType.Value == "call_forwarding"
		})
		if !ok {
			if onDelete {
				return nil
			}
			return fmt.Errorf("call_fowarding not found on business hours")
		}
		for _, setting := range callForwarding.Settings.Value.CallForwardingSettings {
			desc := setting.Description.Value
			if desc == callForwardingDescriptionZoomMobileApps {
				callForwardingSettings = append(callForwardingSettings, zoomphone.PatchCallHandlingSettingsCallForwardingSettingsCallForwardingSettingsItem{
					ID:     setting.ID,
					Enable: util.ToPhoneOptBool(dto.enableZoomMobileApps),
				})
			} else if desc == callForwardingDescriptionZoomDesktopApps {
				callForwardingSettings = append(callForwardingSettings, zoomphone.PatchCallHandlingSettingsCallForwardingSettingsCallForwardingSettingsItem{
					ID:     setting.ID,
					Enable: util.ToPhoneOptBool(dto.enableZoomDesktopApps),
				})
			} else if desc == callForwardingDescriptionZoomPhoneApplianceApps {
				callForwardingSettings = append(callForwardingSettings, zoomphone.PatchCallHandlingSettingsCallForwardingSettingsCallForwardingSettingsItem{
					ID:     setting.ID,
					Enable: util.ToPhoneOptBool(dto.enableZoomPhoneApplianceApps),
				})
			}
		}
		break
	case settingTypeClosedHours:
		callForwarding, ok := lo.Find(detail.ClosedHours, func(item zoomphone.GetCallHandlingOKClosedHoursItem) bool {
			return item.SubSettingType.Value == "call_forwarding"
		})
		if !ok {
			if onDelete {
				return nil
			}
			return fmt.Errorf("call_fowarding not found on closed hours")
		}
		for _, setting := range callForwarding.Settings.Value.CallForwardingSettings {
			desc := setting.Description.Value
			if desc == callForwardingDescriptionZoomMobileApps {
				callForwardingSettings = append(callForwardingSettings, zoomphone.PatchCallHandlingSettingsCallForwardingSettingsCallForwardingSettingsItem{
					ID:     setting.ID,
					Enable: util.ToPhoneOptBool(dto.enableZoomMobileApps),
				})
			} else if desc == callForwardingDescriptionZoomDesktopApps {
				callForwardingSettings = append(callForwardingSettings, zoomphone.PatchCallHandlingSettingsCallForwardingSettingsCallForwardingSettingsItem{
					ID:     setting.ID,
					Enable: util.ToPhoneOptBool(dto.enableZoomDesktopApps),
				})
			} else if desc == callForwardingDescriptionZoomPhoneApplianceApps {
				callForwardingSettings = append(callForwardingSettings, zoomphone.PatchCallHandlingSettingsCallForwardingSettingsCallForwardingSettingsItem{
					ID:     setting.ID,
					Enable: util.ToPhoneOptBool(dto.enableZoomPhoneApplianceApps),
				})
			}
		}
		break
	case settingTypeHolidayHours:
		holiday, ok := lo.Find(detail.HolidayHours, func(item zoomphone.GetCallHandlingOKHolidayHoursItem) bool {
			return item.HolidayID.Value == dto.holidayID.ValueString()
		})
		if !ok {
			if onDelete {
				return nil
			}
			return fmt.Errorf("holiday setting not found on holiday hours")
		}
		callForwarding, ok := lo.Find(holiday.Details, func(item zoomphone.GetCallHandlingOKHolidayHoursItemDetailsItem) bool {
			return item.SubSettingType.Value == "call_forwarding"
		})
		if !ok {
			if onDelete {
				return nil
			}
			return fmt.Errorf("call_fowarding not found on holiday hours")
		}
		for _, setting := range callForwarding.Settings.Value.CallForwardingSettings {
			desc := setting.Description.Value
			if desc == callForwardingDescriptionZoomMobileApps {
				callForwardingSettings = append(callForwardingSettings, zoomphone.PatchCallHandlingSettingsCallForwardingSettingsCallForwardingSettingsItem{
					ID:     setting.ID,
					Enable: util.ToPhoneOptBool(dto.enableZoomMobileApps),
				})
			} else if desc == callForwardingDescriptionZoomDesktopApps {
				callForwardingSettings = append(callForwardingSettings, zoomphone.PatchCallHandlingSettingsCallForwardingSettingsCallForwardingSettingsItem{
					ID:     setting.ID,
					Enable: util.ToPhoneOptBool(dto.enableZoomDesktopApps),
				})
			} else if desc == callForwardingDescriptionZoomPhoneApplianceApps {
				callForwardingSettings = append(callForwardingSettings, zoomphone.PatchCallHandlingSettingsCallForwardingSettingsCallForwardingSettingsItem{
					ID:     setting.ID,
					Enable: util.ToPhoneOptBool(dto.enableZoomPhoneApplianceApps),
				})
			}
		}
		break
	default:
		return fmt.Errorf("unknown setting type, provider implementation error: %s", dto.settingType)
	}

	for _, item := range dto.settings {
		callForwardingSettings = append(callForwardingSettings, zoomphone.PatchCallHandlingSettingsCallForwardingSettingsCallForwardingSettingsItem{
			Description: util.ToPhoneOptString(item.description),
			Enable:      util.ToPhoneOptBool(item.enable),
			ID:          util.ToPhoneOptString(item.id),
			PhoneNumber: util.ToPhoneOptString(item.phoneNumber),
			ExternalContact: zoomphone.OptPatchCallHandlingSettingsCallForwardingSettingsCallForwardingSettingsItemExternalContact{
				Value: zoomphone.PatchCallHandlingSettingsCallForwardingSettingsCallForwardingSettingsItemExternalContact{
					ExternalContactID: util.ToPhoneOptString(item.externalContactID),
				},
				Set: true,
			},
		})
	}
	err = c.client.UpdateCallHandling(ctx, zoomphone.OptUpdateCallHandlingReq{
		Value: zoomphone.UpdateCallHandlingReq{
			Type: zoomphone.PatchCallHandlingSettingsCallForwardingUpdateCallHandlingReq,
			PatchCallHandlingSettingsCallForwarding: zoomphone.PatchCallHandlingSettingsCallForwarding{
				Settings: zoomphone.OptPatchCallHandlingSettingsCallForwardingSettings{
					Value: zoomphone.PatchCallHandlingSettingsCallForwardingSettings{
						HolidayID:                     util.ToPhoneOptString(dto.holidayID),
						RequirePress1BeforeConnecting: util.ToPhoneOptBool(dto.requirePress1BeforeConnecting),
						CallForwardingSettings:        callForwardingSettings,
					},
					Set: true,
				},
				SubSettingType: zoomphone.NewOptString("call_forwarding"),
			},
		},
		Set: true,
	}, zoomphone.UpdateCallHandlingParams{
		ExtensionId: dto.extensionID.ValueString(),
		SettingType: string(dto.settingType),
	})
	if err != nil {
		return fmt.Errorf("error patching phone call handling on call forwarding: %v", err)
	}

	return nil
}

func (c *crud) deleteCallForwarding(ctx context.Context, dto *deleteCallForwardingDto) error {
	err := c.client.DeleteCallHandling(ctx, zoomphone.DeleteCallHandlingParams{
		ExtensionId:      dto.extensionID.ValueString(),
		SettingType:      string(dto.settingType),
		CallForwardingID: util.ToPhoneOptString(dto.callForwardingID),
	})
	if err != nil {
		var status *zoomphone.ErrorResponseStatusCode
		if errors.As(err, &status) {
			if status.StatusCode == 404 && status.Response.Code.Value == 404 {
				return nil
			}
		}
		return fmt.Errorf("error deleting phone call handling on call forwarding: %v", err)
	}

	return nil
}

func (c *crud) deleteHoliday(ctx context.Context, dto *deleteHolidayDto) error {
	err := c.client.DeleteCallHandling(ctx, zoomphone.DeleteCallHandlingParams{
		ExtensionId: dto.extensionID.ValueString(),
		SettingType: string(dto.settingType),
		HolidayID:   util.ToPhoneOptString(dto.holidayID),
	})
	if err != nil {
		var status *zoomphone.ErrorResponseStatusCode
		if errors.As(err, &status) {
			if status.StatusCode == 404 && status.Response.Code.Value == 404 {
				return nil
			}
		}
		return fmt.Errorf("error deleting phone call handling on holiday: %v", err)
	}

	return nil
}
