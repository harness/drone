// Copyright 2022 Harness Inc. All rights reserved.
// Use of this source code is governed by the Polyform Free Trial License
// that can be found in the LICENSE.md file for this repository.

package pullreq

import (
	"context"
	"fmt"

	"github.com/harness/gitness/internal/auth"
	"github.com/harness/gitness/types/enum"
)

// ReviewerDelete deletes reviewer from the reviewerlist for the given PR.
func (c *Controller) ReviewerDelete(ctx context.Context, session *auth.Session,
	repoRef string, prID, reviewerID int64) error {
	_, err := c.getRepoCheckAccess(ctx, session, repoRef, enum.PermissionRepoEdit)
	if err != nil {
		return fmt.Errorf("failed to acquire access to repo: %w", err)
	}

	err = c.reviewerStore.Delete(ctx, prID, reviewerID)
	if err != nil {
		return fmt.Errorf("failed to delete reviewer: %w", err)
	}
	return nil
}
