package usercallingplans

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"github.com/folio-sec/terraform-provider-zoom/generated/api/zoomphone"
	"github.com/folio-sec/terraform-provider-zoom/internal/util"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/samber/lo"
	lop "github.com/samber/lo/parallel"
)

func newCrud(client *zoomphone.Client) *crud {
	return &crud{
		client: client,
	}
}

type crud struct {
	client *zoomphone.Client
}

func (c *crud) read(ctx context.Context, userID types.String) (*readDto, error) {
	ret, err := c.client.PhoneUser(ctx, zoomphone.PhoneUserParams{
		UserId: userID.ValueString(),
	})
	if err != nil {
		return nil, fmt.Errorf("unable to read phone user: %v", err)
	}

	return &readDto{
		callingPlans: lo.Map(ret.CallingPlans, func(v zoomphone.PhoneUserOKCallingPlansItem, _ int) readDtoCallingPlan {
			return readDtoCallingPlan{
				callingPlanType:  util.FromOptInt(v.Type),
				callingPlanName:  util.FromOptString(v.Name),
				billingAccountID: util.FromOptString(v.BillingAccountID),
			}
		}),
	}, nil
}

func (c *crud) create(ctx context.Context, dto createDto) (*createdDto, error) {
	err := c.client.AssignCallingPlan(ctx, zoomphone.NewOptAssignCallingPlanReq(zoomphone.AssignCallingPlanReq{
		CallingPlans: lo.Map(dto.callingPlans, func(v createDtoCallingPlan, _ int) zoomphone.AssignCallingPlanReqCallingPlansItem {
			return zoomphone.AssignCallingPlanReqCallingPlansItem{
				Type:             util.ToPhoneOptInt(v.callingPlanType),
				BillingAccountID: util.ToPhoneOptString(v.billingAccountID),
			}
		}),
	}), zoomphone.AssignCallingPlanParams{
		UserId: dto.userID.ValueString(),
	})
	if err != nil {
		return nil, fmt.Errorf("unable to create phone user calling plan: %v", err)
	}
	return &createdDto{}, nil
}

func (c *crud) delete(ctx context.Context, dto deleteDto) error {
	errs := lo.Compact(lop.Map(dto.callingPlans, func(v deleteDtoCallingPlan, _ int) error {
		err := c.client.UnassignCallingPlan(ctx, zoomphone.UnassignCallingPlanParams{
			UserId:           dto.userID.ValueString(),
			PlanType:         strconv.Itoa(int(v.callingPlanType.ValueInt32())),
			BillingAccountID: util.ToPhoneOptString(v.billingAccountID),
		})

		if err != nil {
			var status *zoomphone.ErrorResponseStatusCode
			if errors.As(err, &status) {
				if status.StatusCode == 404 {
					return nil // already deleted
				}
			}
			return fmt.Errorf("unable to delete phone user calling plan: %v", err)
		}

		return nil
	}))

	if len(errs) > 0 {
		return errors.Join(errs...)
	}

	return nil
}
