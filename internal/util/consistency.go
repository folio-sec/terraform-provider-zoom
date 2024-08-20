// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package util

// Because there is no alternative available for terraform-plugin-framework, terraform-plugin-sdk/v2 is unavoidably used.
// See also: https://github.com/hashicorp/terraform-plugin-framework/issues/513
// See also: https://discuss.hashicorp.com/t/terraform-plugin-framework-what-is-the-replacement-for-waitforstate-or-retrycontext/45538

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
)

type ChangeFunc func(ctx context.Context) (*bool, error)

func WaitForDeletion(ctx context.Context, f ChangeFunc) error {
	deadline, ok := ctx.Deadline()
	if !ok {
		return errors.New("context has no deadline")
	}

	timeout := time.Until(deadline)
	_, err := (&retry.StateChangeConf{ //nolint:staticcheck
		Pending:                   []string{"Waiting"},
		Target:                    []string{"Deleted"},
		Timeout:                   timeout,
		MinTimeout:                5 * time.Second,
		ContinuousTargetOccurence: 5,
		Refresh: func() (interface{}, string, error) {
			exists, err := f(ctx)
			if err != nil {
				return nil, "Error", fmt.Errorf("retrieving resource: %+v", err)
			}
			if exists == nil {
				return nil, "Error", fmt.Errorf("retrieving resource via ChangeFunc returned nil")
			}
			if *exists {
				return "stub", "Waiting", nil
			}
			return "stub", "Deleted", nil
		},
	}).WaitForStateContext(ctx)

	return err
}

func WaitForUpdate(ctx context.Context, f ChangeFunc) error {
	deadline, ok := ctx.Deadline()
	if !ok {
		return errors.New("context has no deadline")
	}

	_, err := WaitForUpdateWithTimeout(ctx, time.Until(deadline), f)
	return err
}

func WaitForUpdateWithTimeout(ctx context.Context, timeout time.Duration, f ChangeFunc) (bool, error) {
	res, err := (&retry.StateChangeConf{ //nolint:staticcheck
		Pending:                   []string{"Waiting"},
		Target:                    []string{"Done"},
		Timeout:                   timeout,
		MinTimeout:                5 * time.Second,
		ContinuousTargetOccurence: 5,
		Refresh: func() (interface{}, string, error) {
			updated, err := f(ctx)
			if err != nil {
				return nil, "Error", fmt.Errorf("retrieving resource: %+v", err)
			}
			if updated == nil {
				return nil, "Error", fmt.Errorf("retrieving resource via ChangeFunc returned nil")
			}
			if *updated {
				return true, "Done", nil
			}
			return false, "Waiting", nil
		},
	}).WaitForStateContext(ctx)

	if res == nil {
		return false, err
	}
	return res.(bool), err
}
