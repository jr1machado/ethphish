package models

import (
	"time"

	check "gopkg.in/check.v1"
)

// TestPortalLoginTokenRoundTrip verifies a self-service login token can be
// created, redeemed once, and that a second redemption of the same token
// fails — it's single-use even before it expires.
func (s *ModelsSuite) TestPortalLoginTokenRoundTrip(c *check.C) {
	campaign := s.createCampaignDependencies(c)
	contract := Contract{Name: "Portal Login Contract", ClientName: "Acme"}
	err := PostContractForTenant(&contract, 1, campaign.UserId)
	c.Assert(err, check.Equals, nil)
	err = AddContractApprover(&ContractApprover{ContractId: contract.Id, Name: "Portal Approver", Email: "portal.approver@example.com"})
	c.Assert(err, check.Equals, nil)

	token, err := CreatePortalLoginToken(1, "portal.approver@example.com", 15*time.Minute)
	c.Assert(err, check.Equals, nil)

	plt, err := RedeemPortalLoginToken(token)
	c.Assert(err, check.Equals, nil)
	c.Assert(plt.Email, check.Equals, "portal.approver@example.com")
	c.Assert(plt.TenantID, check.Equals, int64(1))

	_, err = RedeemPortalLoginToken(token)
	c.Assert(err, check.Equals, ErrPortalLoginTokenInvalid)
}

// TestPortalLoginTokenExpired verifies an expired token is rejected even
// though it was never redeemed.
func (s *ModelsSuite) TestPortalLoginTokenExpired(c *check.C) {
	token, err := CreatePortalLoginToken(1, "expired@example.com", -1*time.Minute)
	c.Assert(err, check.Equals, nil)

	_, err = RedeemPortalLoginToken(token)
	c.Assert(err, check.Equals, ErrPortalLoginTokenInvalid)
}

// TestPortalLoginTokenWrongToken verifies a token that was never issued is
// rejected.
func (s *ModelsSuite) TestPortalLoginTokenWrongToken(c *check.C) {
	_, err := RedeemPortalLoginToken("not-a-real-token")
	c.Assert(err, check.Equals, ErrPortalLoginTokenInvalid)
}

// TestFindTenantsForApproverEmail verifies the self-service login lookup
// resolves the tenant from a registered approver's e-mail, and returns
// nothing for an e-mail that isn't a configured approver anywhere.
func (s *ModelsSuite) TestFindTenantsForApproverEmail(c *check.C) {
	campaign := s.createCampaignDependencies(c)
	contract := Contract{Name: "Lookup Contract", ClientName: "Acme"}
	err := PostContractForTenant(&contract, 1, campaign.UserId)
	c.Assert(err, check.Equals, nil)
	err = AddContractApprover(&ContractApprover{ContractId: contract.Id, Name: "Lookup Approver", Email: "lookup.approver@example.com"})
	c.Assert(err, check.Equals, nil)

	matches, err := FindTenantsForApproverEmail("lookup.approver@example.com")
	c.Assert(err, check.Equals, nil)
	c.Assert(len(matches), check.Equals, 1)
	c.Assert(matches[0].TenantID, check.Equals, int64(1))
	c.Assert(matches[0].Name, check.Equals, "Lookup Approver")

	none, err := FindTenantsForApproverEmail("nobody@example.com")
	c.Assert(err, check.Equals, nil)
	c.Assert(len(none), check.Equals, 0)
}
