package models

import "gopkg.in/check.v1"

// TestTemplateRoundTripCharacterization preserves the template behavior relied
// upon by campaigns before the persistence layer is moved to PostgreSQL.
func (s *ModelsSuite) TestTemplateRoundTripCharacterization(c *check.C) {
	template := Template{
		UserId:  1,
		Name:    "Baseline template",
		Subject: "Welcome {{.FirstName}}",
		Text:    "Hello {{.FirstName}}",
		HTML:    "<p>Hello {{.FirstName}}</p>",
	}
	c.Assert(PostTemplate(&template), check.IsNil)
	c.Assert(template.Id, check.Not(check.Equals), int64(0))

	loaded, err := GetTemplate(template.Id, template.UserId)
	c.Assert(err, check.IsNil)
	c.Assert(loaded.Name, check.Equals, template.Name)
	c.Assert(loaded.Subject, check.Equals, template.Subject)
	c.Assert(loaded.HTML, check.Equals, template.HTML)

	_, err = GetTemplate(template.Id, 2)
	c.Assert(err, check.NotNil)
}

func (s *ModelsSuite) TestTemplateValidationCharacterization(c *check.C) {
	c.Assert((&Template{}).Validate(), check.Equals, ErrTemplateNameNotSpecified)
	c.Assert((&Template{Name: "Empty"}).Validate(), check.Equals, ErrTemplateMissingParameter)
}
