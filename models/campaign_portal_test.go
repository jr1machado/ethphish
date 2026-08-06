package models

import (
	"time"

	check "gopkg.in/check.v1"
)

// TestCampaignSummariesForTenantAllUsersIsTenantScoped verifies the
// client-portal campaign listing only returns campaigns for the requested
// tenant, and nothing for a tenant with none — the same isolation
// guarantee as the admin-facing GetCampaignSummariesForTenant, just
// without the additional per-admin-user filter.
func (s *ModelsSuite) TestCampaignSummariesForTenantAllUsersIsTenantScoped(c *check.C) {
	// Inserted directly rather than via PostCampaignForTenant, which also
	// launches the campaign (mail dispatch) — this test only needs a
	// campaigns row to list, not a real send.
	campaign := Campaign{Name: "Portal Test Campaign", UserId: 1, TenantID: 1, Status: CampaignQueued, CreatedDate: time.Now().UTC(), LaunchDate: time.Now().UTC()}
	err := db.Save(&campaign).Error
	c.Assert(err, check.Equals, nil)

	summaries, err := GetCampaignSummariesForTenantAllUsers(1)
	c.Assert(err, check.Equals, nil)
	c.Assert(len(summaries.Campaigns) >= 1, check.Equals, true)

	other := Tenant{Name: "Other Tenant", Slug: "other-tenant"}
	err = PostTenant(&other)
	c.Assert(err, check.Equals, nil)

	empty, err := GetCampaignSummariesForTenantAllUsers(other.ID)
	c.Assert(err, check.Equals, nil)
	c.Assert(len(empty.Campaigns), check.Equals, 0)
}
