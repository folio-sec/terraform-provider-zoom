package autoreceptionistivr

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"github.com/folio-sec/terraform-provider-zoom/generated/api/zoomphone"
	"github.com/folio-sec/terraform-provider-zoom/internal/util"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func newCrud(client *zoomphone.Client) *crud {
	return &crud{
		client: client,
	}
}

type crud struct {
	client *zoomphone.Client
}

func (c *crud) read(ctx context.Context, autoReceptionistID, hoursType, holidayID types.String) (*readDto, error) {
	detail, err := c.client.GetAutoReceptionistIVR(ctx, zoomphone.GetAutoReceptionistIVRParams{
		AutoReceptionistId: autoReceptionistID.ValueString(),
		HoursType:          util.ToPhoneOptString(hoursType),
		HolidayID:          util.ToPhoneOptString(holidayID),
	})
	if err != nil {
		var status *zoomphone.ErrorResponseStatusCode
		if errors.As(err, &status) {
			if status.StatusCode == 400 && status.Response.Code.Value == 300 {
				return nil, nil // already deleted
			}
		}
		return nil, fmt.Errorf("unable to read phone auto receptionist ivr: %w", err)
	}

	var audioPrompt *readDtoAudioPrompt
	if detail.AudioPrompt.IsSet() {
		audioPrompt = &readDtoAudioPrompt{
			id:   util.FromOptString(detail.AudioPrompt.Value.ID),
			name: util.FromOptString(detail.AudioPrompt.Value.Name),
		}
	}
	var callerEntersNoAction *readDtoCallerEntersNoAction
	if detail.CallerEntersNoAction.IsSet() {
		var forwardTo *readDtoCallerEntersNoActionForwardTo
		if detail.CallerEntersNoAction.Value.ForwardTo.IsSet() {
			forwardTo = &readDtoCallerEntersNoActionForwardTo{
				displayName:     util.FromOptString(detail.CallerEntersNoAction.Value.ForwardTo.Value.DisplayName),
				extensionID:     util.FromOptString(detail.CallerEntersNoAction.Value.ForwardTo.Value.ExtensionID),
				extensionNumber: util.FromOptString(detail.CallerEntersNoAction.Value.ForwardTo.Value.ExtensionNumber),
				id:              util.FromOptString(detail.CallerEntersNoAction.Value.ForwardTo.Value.ID),
			}
		}
		callerEntersNoAction = &readDtoCallerEntersNoAction{
			action:            util.FromOptInt(detail.CallerEntersNoAction.Value.Action),
			audioPromptRepeat: util.FromOptInt(detail.CallerEntersNoAction.Value.AudioPromptRepeat),
			forwardTo:         forwardTo,
		}
	}
	var keyActions []*readDtoKeyAction
	for _, keyAction := range detail.KeyActions {
		var target *readDtoKeyActionTarget
		if keyAction.Target.IsSet() {
			target = &readDtoKeyActionTarget{
				displayName:     util.FromOptString(keyAction.Target.Value.DisplayName),
				extensionID:     util.FromOptString(keyAction.Target.Value.ExtensionID),
				extensionNumber: util.FromOptString(keyAction.Target.Value.ExtensionNumber),
				id:              util.FromOptString(keyAction.Target.Value.ID),
				phoneNumber:     util.FromOptString(keyAction.Target.Value.PhoneNumber),
			}
		}
		var voicemailGreeting *readDtoKeyActionVoicemailGreeting
		if keyAction.VoicemailGreeting.IsSet() {
			voicemailGreeting = &readDtoKeyActionVoicemailGreeting{
				id:   util.FromOptString(keyAction.VoicemailGreeting.Value.ID),
				name: util.FromOptString(keyAction.VoicemailGreeting.Value.Name),
			}
		}
		keyActions = append(keyActions, &readDtoKeyAction{
			action:            util.FromOptInt(keyAction.Action),
			key:               util.FromOptString(keyAction.Key),
			target:            target,
			voicemailGreeting: voicemailGreeting,
		})
	}
	return &readDto{
		autoReceptionistID:   autoReceptionistID,
		hoursType:            hoursType,
		holidayID:            holidayID,
		audioPrompt:          audioPrompt,
		callerEntersNoAction: callerEntersNoAction,
		keyActions:           keyActions,
	}, nil
}

func (c *crud) update(ctx context.Context, dto *updateDto) error {
	// zoom update api allows only one keyAction, so update process do followings
	// 1. update fields except keyActions by (auto_reception_id x holiday_id x hours_type)
	// 2. update keyActions by (auto_reception_id x holiday_id x hours_type)

	callerEntersNoAction := zoomphone.OptUpdateAutoReceptionistIVRReqCallerEntersNoAction{}
	if dto.callerEntersNoAction != nil {
		callerEntersNoAction = zoomphone.NewOptUpdateAutoReceptionistIVRReqCallerEntersNoAction(zoomphone.UpdateAutoReceptionistIVRReqCallerEntersNoAction{
			Action:               util.ToPhoneOptInt(dto.callerEntersNoAction.action),
			AudioPromptRepeat:    util.ToPhoneOptInt(dto.callerEntersNoAction.auditPromptRepeat),
			ForwardToExtensionID: util.ToPhoneOptString(dto.callerEntersNoAction.forwardToExtensionID),
		})
	}
	err := c.client.UpdateAutoReceptionistIVR(ctx, zoomphone.OptUpdateAutoReceptionistIVRReq{
		Value: zoomphone.UpdateAutoReceptionistIVRReq{
			HolidayID:            util.ToPhoneOptString(dto.holidayID),
			HoursType:            util.ToPhoneOptString(dto.hoursType),
			AudioPromptID:        util.ToPhoneOptString(dto.audioPromptID),
			CallerEntersNoAction: callerEntersNoAction,
		},
		Set: true,
	}, zoomphone.UpdateAutoReceptionistIVRParams{
		AutoReceptionistId: dto.autoReceptionistID.ValueString(),
	})
	if err != nil {
		return fmt.Errorf("error updating phone auto receptionist ivr: %v", err)
	}

	for _, keyAction := range dto.keyActions {
		target := zoomphone.OptUpdateAutoReceptionistIVRReqKeyActionTarget{}
		if keyAction.target != nil {
			target = zoomphone.NewOptUpdateAutoReceptionistIVRReqKeyActionTarget(zoomphone.UpdateAutoReceptionistIVRReqKeyActionTarget{
				ExtensionID: util.ToPhoneOptString(keyAction.target.extensionID),
				PhoneNumber: util.ToPhoneOptString(keyAction.target.phoneNumber),
			})
		}
		err := c.client.UpdateAutoReceptionistIVR(ctx, zoomphone.OptUpdateAutoReceptionistIVRReq{
			Value: zoomphone.UpdateAutoReceptionistIVRReq{
				HolidayID: util.ToPhoneOptString(dto.holidayID),
				HoursType: util.ToPhoneOptString(dto.hoursType),
				KeyAction: zoomphone.NewOptUpdateAutoReceptionistIVRReqKeyAction(zoomphone.UpdateAutoReceptionistIVRReqKeyAction{
					Key:                 util.ToPhoneOptString(keyAction.key),
					Action:              util.ToPhoneOptInt(keyAction.action),
					Target:              target,
					VoicemailGreetingID: util.ToPhoneOptString(keyAction.voiceMailGreetingId),
				}),
			},
			Set: true,
		}, zoomphone.UpdateAutoReceptionistIVRParams{
			AutoReceptionistId: dto.autoReceptionistID.ValueString(),
		})
		if err != nil {
			return fmt.Errorf("error updating phone auto receptionist ivr on key=%s: %v", keyAction.key, err)
		}
	}
	return nil
}

func (c *crud) delete(ctx context.Context, autoReceptionistID, hoursType, holidayID types.String) error {
	// there is no delete api, so just update with following initial values
	// - AudioPromptID
	//   - Default (id is empty string)
	// - CallerEntersNoAction
	//   - Disabled
	// - KeyActions
	//   - 0: Leave Voicemail to Current Extension
	//   - 1-9: Disabled
	//   - *: Repeat menu greeting
	//   - #: Disabled

	// update fields with initial values as deleting process
	var keyActions []*updateDtoKeyAction
	keyActions = append(keyActions,
		&updateDtoKeyAction{
			key:                 types.StringValue("0"),
			action:              types.Int32Value(100), // 100 Leave voicemail to the current extension
			target:              nil,
			voiceMailGreetingId: types.StringValue(""), // default
		},
		&updateDtoKeyAction{
			key:    types.StringValue("*"),
			action: types.Int32Value(21), // 21 Repeat menu greeting
		},
		&updateDtoKeyAction{
			key:    types.StringValue("#"),
			action: types.Int32Value(-1), // -1 Disabled
		},
	)
	for i := 1; i < 10; i++ {
		keyActions = append(keyActions, &updateDtoKeyAction{
			key:    types.StringValue(strconv.Itoa(i)),
			action: types.Int32Value(-1), // -1 Disabled
		})
	}
	err := c.update(ctx, &updateDto{
		autoReceptionistID: autoReceptionistID,
		holidayID:          holidayID,
		hoursType:          hoursType,
		audioPromptID:      types.StringValue(""), // default
		callerEntersNoAction: &updateDtoCallerEntersNoAction{
			action:            types.Int32Value(-1), // -1 Disabled
			auditPromptRepeat: types.Int32Value(3),
		},
		keyActions: keyActions,
	})
	if err != nil {
		return fmt.Errorf("error deleting phone auto receptionist ivr: %v", err)
	}
	return nil
}
