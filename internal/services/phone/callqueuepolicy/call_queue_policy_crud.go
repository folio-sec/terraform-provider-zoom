package callqueuepolicy

import (
	"context"
	"errors"
	"fmt"

	"github.com/folio-sec/terraform-provider-zoom/generated/api/zoomphone"
	"github.com/folio-sec/terraform-provider-zoom/internal/util"
	"github.com/hashicorp/terraform-plugin-framework/types"
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

func (c *crud) read(ctx context.Context, callQueueID types.String) (*readDto, error) {
	detail, err := c.client.GetACallQueue(ctx, zoomphone.GetACallQueueParams{
		CallQueueId: callQueueID.ValueString(),
	})
	if err != nil {
		var status *zoomphone.ErrorResponseStatusCode
		if errors.As(err, &status) {
			if status.StatusCode == 400 {
				return nil, nil // already deleted
			}
		}
		return nil, fmt.Errorf("unable to read phone call queue policy: %w", err)
	}

	var policyVoiceMailMembers []*readDtoPolicyVoiceMailMember
	if detail.Policy.IsSet() {
		policyVoiceMailMembers = lo.Map(detail.Policy.Value.GetVoicemailAccessMembers(), func(item zoomphone.GetACallQueueOKPolicyVoicemailAccessMembersItem, index int) *readDtoPolicyVoiceMailMember {
			return &readDtoPolicyVoiceMailMember{
				accessUserID:  util.FromOptString(item.AccessUserID),
				allowDownload: util.FromOptBool(item.AllowDownload),
				allowDelete:   util.FromOptBool(item.AllowDelete),
				allowSharing:  util.FromOptBool(item.AllowSharing),
				sharedID:      util.FromOptString(item.SharedID),
			}
		})
	}
	return &readDto{
		callQueueID:            callQueueID,
		policyVoiceMailMembers: policyVoiceMailMembers,
	}, nil
}

func (c *crud) add(ctx context.Context, dto *addDto) error {
	var voicemailAccessMembers []zoomphone.AddCQPolicySubSettingReqVoicemailAccessMembersItem
	if dto.voicemailAccessMembers != nil {
		voicemailAccessMembers = lo.Map(dto.voicemailAccessMembers, func(item *addDtoVoicemailAccessMember, index int) zoomphone.AddCQPolicySubSettingReqVoicemailAccessMembersItem {
			return zoomphone.AddCQPolicySubSettingReqVoicemailAccessMembersItem{
				AccessUserID:  util.ToPhoneOptString(item.accessUserID),
				AllowDownload: util.ToPhoneOptBool(item.allowDownload),
				AllowDelete:   util.ToPhoneOptBool(item.allowDelete),
				AllowSharing:  util.ToPhoneOptBool(item.allowSharing),
			}
		})
	}
	_, err := c.client.AddCQPolicySubSetting(ctx, zoomphone.OptAddCQPolicySubSettingReq{
		Value: zoomphone.AddCQPolicySubSettingReq{
			VoicemailAccessMembers: voicemailAccessMembers,
		},
		Set: true,
	}, zoomphone.AddCQPolicySubSettingParams{
		CallQueueId: dto.callQueueID.ValueString(),
		PolicyType:  dto.policyType.String(),
	})
	if err != nil {
		return fmt.Errorf("error creating phone call queue policy: %v", err)
	}
	return nil
}

func (c *crud) update(ctx context.Context, dto *updateDto) error {
	var voicemailAccessMembers []zoomphone.UpdateCQPolicySubSettingReqVoicemailAccessMembersItem
	if dto.voicemailAccessMembers != nil {
		voicemailAccessMembers = lo.Map(dto.voicemailAccessMembers, func(item *updateDtoVoicemailAccessMember, index int) zoomphone.UpdateCQPolicySubSettingReqVoicemailAccessMembersItem {
			return zoomphone.UpdateCQPolicySubSettingReqVoicemailAccessMembersItem{
				AccessUserID:  util.ToPhoneOptString(item.accessUserID),
				AllowDownload: util.ToPhoneOptBool(item.allowDownload),
				AllowDelete:   util.ToPhoneOptBool(item.allowDelete),
				AllowSharing:  util.ToPhoneOptBool(item.allowSharing),
				SharedID:      util.ToPhoneOptString(item.sharedID),
			}
		})
	}
	// Due to the patch API specifications, an error occurs as 'voicemail_access_members cannot be empty' with using emtpy list.
	// Therefore, if the slice is empty, the function returns nil without performing any actions.
	if len(voicemailAccessMembers) == 0 {
		return nil
	}

	err := c.client.UpdateCQPolicySubSetting(ctx, zoomphone.OptUpdateCQPolicySubSettingReq{
		Value: zoomphone.UpdateCQPolicySubSettingReq{
			VoicemailAccessMembers: voicemailAccessMembers,
		},
		Set: true,
	}, zoomphone.UpdateCQPolicySubSettingParams{
		CallQueueId: dto.callQueueID.ValueString(),
		PolicyType:  dto.policyType.String(),
	})
	if err != nil {
		return fmt.Errorf("error updating phone call queue policy: %v", err)
	}
	return nil
}

func (c *crud) remove(ctx context.Context, dto *removeDto) error {
	// maxItems: 20
	for _, chunk := range lo.Chunk(dto.sharedIDs, 20) {
		err := c.client.RemoveCQPolicySubSetting(ctx, zoomphone.RemoveCQPolicySubSettingParams{
			CallQueueId: dto.callQueueID.ValueString(),
			PolicyType:  dto.policyType.String(),
			SharedIds: lo.Map(chunk, func(item types.String, index int) string {
				return item.ValueString()
			}),
		})
		if err != nil {
			var status *zoomphone.ErrorResponseStatusCode
			if errors.As(err, &status) {
				if status.StatusCode == 400 && status.Response.Code.Value == 404 {
					return nil
				}
			}
			return fmt.Errorf("error removing phone call queue policy: %v", err)
		}
	}
	return nil
}
