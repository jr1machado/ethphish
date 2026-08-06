package models

import (
	"time"

	check "gopkg.in/check.v1"
)

func (s *ModelsSuite) createApprovedContract(c *check.C) (Campaign, *int64) {
	campaign := s.createCampaignDependencies(c)

	contract := Contract{Name: "MSA 2026", ClientName: "Acme Corp"}
	err := PostContractForTenant(&contract, 1, campaign.UserId)
	c.Assert(err, check.Equals, nil)

	version := ContractVersion{FilePath: "/tmp/does-not-matter.pdf", FileName: "msa.pdf", ScopeDescription: "Phase 1"}
	_, err = AddContractVersion(contract.Id, &version, campaign.UserId)
	c.Assert(err, check.Equals, nil)

	ar, _, err := CreateApprovalRequest(version.Id, nil, 24*time.Hour)
	c.Assert(err, check.Equals, nil)
	err = DecideApproval(ar.Id, 1, ApprovalApproved, "looks good")
	c.Assert(err, check.Equals, nil)

	campaign.ContractID = &contract.Id
	return campaign, &contract.Id
}

// TestCampaignBlockedWithoutApproval verifies a campaign linked to a
// contract with no approval at all is rejected.
func (s *ModelsSuite) TestCampaignBlockedWithoutApproval(c *check.C) {
	campaign := s.createCampaignDependencies(c)
	contract := Contract{Name: "Unapproved Contract", ClientName: "Acme"}
	err := PostContractForTenant(&contract, 1, campaign.UserId)
	c.Assert(err, check.Equals, nil)
	version := ContractVersion{FilePath: "/tmp/x.pdf", FileName: "x.pdf"}
	_, err = AddContractVersion(contract.Id, &version, campaign.UserId)
	c.Assert(err, check.Equals, nil)

	campaign.ContractID = &contract.Id
	err = campaign.Validate()
	c.Assert(err, check.Equals, ErrApprovalRequired)
}

// TestCampaignAllowedWithApproval verifies a campaign linked to a contract
// whose current version is approved passes validation.
func (s *ModelsSuite) TestCampaignAllowedWithApproval(c *check.C) {
	campaign, _ := s.createApprovedContract(c)
	err := campaign.Validate()
	c.Assert(err, check.Equals, nil)
}

// TestCampaignBlockedAfterNewVersion verifies that uploading a new contract
// version after an approval invalidates that approval for gating purposes,
// since the approval is bound to the version it was granted for.
func (s *ModelsSuite) TestCampaignBlockedAfterNewVersion(c *check.C) {
	campaign, contractID := s.createApprovedContract(c)
	c.Assert(campaign.Validate(), check.Equals, nil)

	newVersion := ContractVersion{FilePath: "/tmp/x2.pdf", FileName: "x2.pdf", ScopeDescription: "Phase 2"}
	_, err := AddContractVersion(*contractID, &newVersion, campaign.UserId)
	c.Assert(err, check.Equals, nil)

	err = campaign.Validate()
	c.Assert(err, check.Equals, ErrApprovalRequired)
}

// TestCampaignWithoutContractIsUnaffected verifies contract linkage is
// opt-in: a campaign with no ContractID validates exactly as before.
func (s *ModelsSuite) TestCampaignWithoutContractIsUnaffected(c *check.C) {
	campaign := s.createCampaignDependencies(c)
	c.Assert(campaign.ContractID, check.IsNil)
	err := campaign.Validate()
	c.Assert(err, check.Equals, nil)
}

// TestExpirePendingApprovals verifies the reminder/expiration cron marks
// past-due pending approvals as expired.
func (s *ModelsSuite) TestExpirePendingApprovals(c *check.C) {
	campaign := s.createCampaignDependencies(c)
	contract := Contract{Name: "Expiring Contract", ClientName: "Acme"}
	err := PostContractForTenant(&contract, 1, campaign.UserId)
	c.Assert(err, check.Equals, nil)
	version := ContractVersion{FilePath: "/tmp/e.pdf", FileName: "e.pdf"}
	_, err = AddContractVersion(contract.Id, &version, campaign.UserId)
	c.Assert(err, check.Equals, nil)

	ar, _, err := CreateApprovalRequest(version.Id, nil, -1*time.Hour) // already expired
	c.Assert(err, check.Equals, nil)

	n, err := ExpirePendingApprovals()
	c.Assert(err, check.Equals, nil)
	c.Assert(n >= 1, check.Equals, true)

	updated, err := GetApprovalRequest(ar.Id)
	c.Assert(err, check.Equals, nil)
	c.Assert(updated.Status, check.Equals, ApprovalExpired)
}

// TestTokenRoundTrip verifies a magic-link token can be generated, hashed,
// and matched back to its approval request, and that a wrong token fails.
func (s *ModelsSuite) TestTokenRoundTrip(c *check.C) {
	campaign := s.createCampaignDependencies(c)
	contract := Contract{Name: "Token Contract", ClientName: "Acme"}
	err := PostContractForTenant(&contract, 1, campaign.UserId)
	c.Assert(err, check.Equals, nil)
	version := ContractVersion{FilePath: "/tmp/t.pdf", FileName: "t.pdf"}
	_, err = AddContractVersion(contract.Id, &version, campaign.UserId)
	c.Assert(err, check.Equals, nil)

	ar, token, err := CreateApprovalRequest(version.Id, nil, 24*time.Hour)
	c.Assert(err, check.Equals, nil)

	found, err := GetApprovalRequestByToken(token)
	c.Assert(err, check.Equals, nil)
	c.Assert(found.Id, check.Equals, ar.Id)

	_, err = GetApprovalRequestByToken("not-the-real-token")
	c.Assert(err, check.Equals, ErrApprovalTokenInvalid)
}
