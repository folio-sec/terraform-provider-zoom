package autoreceptionist

import (
	"context"
	"fmt"

	"github.com/folio-sec/terraform-provider-zoom/generated/api/zoomphone"
	"github.com/folio-sec/terraform-provider-zoom/internal/util"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ datasource.DataSource              = &phoneAutoReceptionistDataSource{}
	_ datasource.DataSourceWithConfigure = &phoneAutoReceptionistDataSource{}
)

func newPhoneReceptionistCrud(client *zoomphone.Client) *phoneAutoReceptionistCrud {
	return &phoneAutoReceptionistCrud{
		client: client,
	}
}

type phoneAutoReceptionistCrud struct {
	client *zoomphone.Client
}

func (c *phoneAutoReceptionistCrud) read(ctx context.Context, autoReceptionistID string) (*readDto, error) {
	detail, err := c.client.GetAutoReceptionistDetail(ctx, zoomphone.GetAutoReceptionistDetailParams{
		AutoReceptionistId: autoReceptionistID,
	})
	if err != nil {
		return nil, fmt.Errorf("unable to read phone auto receptionist: %w %T %T", err, err, detail)
	}

	audioPromptLanguage := types.StringNull()
	if detail.AudioPromptLanguage.IsSet() {
		audioPromptLanguage = types.StringValue(string(detail.AudioPromptLanguage.Value))
	}
	return &readDto{
		autoReceptionistID:  types.StringValue(autoReceptionistID),
		costCenter:          util.FromOptString(detail.CostCenter),
		department:          util.FromOptString(detail.Department),
		extensionNumber:     util.FromOptInt64(detail.ExtensionNumber),
		name:                util.FromOptString(detail.Name),
		timezone:            util.FromOptString(detail.Timezone),
		audioPromptLanguage: audioPromptLanguage,
	}, nil
}

func (c *phoneAutoReceptionistCrud) create(ctx context.Context, dto createDto) (*createdDto, error) {
	res, err := c.client.AddAutoReceptionist(ctx, zoomphone.OptAddAutoReceptionistReq{
		Value: zoomphone.AddAutoReceptionistReq{
			Name: dto.name.ValueString(),
		},
		Set: true,
	})
	if err != nil {
		return nil, fmt.Errorf("error creating phone auto receptionist: %v", err)
	}
	if bad, ok := res.(*zoomphone.AddAutoReceptionistBadRequest); ok {
		return nil, fmt.Errorf("error creating phone auto receptionist: bad request %v", bad)
	}
	if created, ok := res.(*zoomphone.AddAutoReceptionistCreated); ok {
		return &createdDto{
			autoReceptionistID: util.FromOptString(created.ID),
			name:               util.FromOptString(created.Name),
			extensionNumber:    util.FromOptInt64(created.ExtensionNumber),
		}, nil
	}
	return nil, fmt.Errorf("error creating phone auto receptionist: invalid implementation %v", res)
}

func (c *phoneAutoReceptionistCrud) update(ctx context.Context, dto updateDto) error {
	audioPromptLanguage := zoomphone.OptUpdateAutoReceptionistReqAudioPromptLanguage{}
	if dto.audioPromptLanguage.ValueString() != "" {
		for _, lang := range zoomphone.UpdateAutoReceptionistReqAudioPromptLanguageJa.AllValues() {
			if string(lang) == dto.audioPromptLanguage.ValueString() {
				audioPromptLanguage = zoomphone.NewOptUpdateAutoReceptionistReqAudioPromptLanguage(lang)
			}
		}
		if !audioPromptLanguage.IsSet() {
			return fmt.Errorf("invalid audio prompt language: %v", dto.audioPromptLanguage.ValueString())
		}
	}
	ret, err := c.client.UpdateAutoReceptionist(ctx, zoomphone.OptUpdateAutoReceptionistReq{
		Value: zoomphone.UpdateAutoReceptionistReq{
			// CostCenter/Department: to remove it, need to pass empty string. not null.
			CostCenter:          zoomphone.NewOptString(util.ToOptString(dto.costCenter).Or("")),
			Department:          zoomphone.NewOptString(util.ToOptString(dto.department).Or("")),
			ExtensionNumber:     util.ToOptInt64(dto.extensionNumber),
			Name:                util.ToOptString(dto.name),
			AudioPromptLanguage: audioPromptLanguage,
			Timezone:            util.ToOptString(dto.timezone),
		},
		Set: true,
	}, zoomphone.UpdateAutoReceptionistParams{
		AutoReceptionistId: dto.autoReceptionistID.ValueString(),
	})
	if err != nil {
		return fmt.Errorf("error updating phone auto receptionist: %v", err)
	}
	if _, ok := ret.(*zoomphone.UpdateAutoReceptionistNoContent); !ok {
		return fmt.Errorf("error updating phone auto receptionist: %v", ret)
	}
	return nil
}

func (c *phoneAutoReceptionistCrud) delete(ctx context.Context, autoReceptionistId string) error {
	ret, err := c.client.DeleteAutoReceptionist(ctx, zoomphone.DeleteAutoReceptionistParams{
		AutoReceptionistId: autoReceptionistId,
	})
	if util.IsUnexpectedStatusCodeError(err, 405) {
		return nil // when passing not exist id, api return 405. but api spec doesn't handle it so just got it as unexpected error
	}
	if err != nil {
		return fmt.Errorf("error deleting phone auto receptionist: %v", err)
	}
	if _, ok := ret.(*zoomphone.DeleteAutoReceptionistNoContent); !ok {
		return fmt.Errorf("error deleting phone auto receptionist: %v, %T", ret, ret)
	}
	return nil
}
