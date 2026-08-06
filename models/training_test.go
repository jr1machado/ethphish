package models

import (
	"strconv"

	check "gopkg.in/check.v1"
)

func (s *ModelsSuite) createTraining(c *check.C, passPercent, maxAttempts int) Training {
	t := Training{Name: "Phishing Awareness"}
	err := PostTrainingForTenant(&t, 1, 1)
	c.Assert(err, check.Equals, nil)

	quiz := TrainingQuiz{TrainingId: t.Id, PassPercent: passPercent, MaxAttempts: maxAttempts}
	err = PostTrainingQuiz(&quiz)
	c.Assert(err, check.Equals, nil)

	q1 := QuizQuestion{QuizId: quiz.Id, Type: QuestionMultipleChoice, Text: "Pick the safe one", Options: `["a","b","c"]`, CorrectAnswer: "b"}
	c.Assert(AddQuizQuestion(&q1), check.Equals, nil)
	q2 := QuizQuestion{QuizId: quiz.Id, Type: QuestionTrueFalse, Text: "Phishing is bad", CorrectAnswer: "true"}
	c.Assert(AddQuizQuestion(&q2), check.Equals, nil)

	full, err := GetTrainingForTenant(t.Id, 1)
	c.Assert(err, check.Equals, nil)
	return full
}

// TestGradeQuizAttemptPassAndFail verifies scoring, case/whitespace
// tolerant answer matching, and the pass/fail threshold.
func (s *ModelsSuite) TestGradeQuizAttemptPassAndFail(c *check.C) {
	training := s.createTraining(c, 80, 0)
	assignment, err := CreateTrainingAssignment(1, training.Id, nil, nil, "target@example.com", "Target")
	c.Assert(err, check.Equals, nil)

	q1, q2 := training.Quiz.Questions[0], training.Quiz.Questions[1]

	// Both correct (case/whitespace-insensitive) -> 100%, passes.
	attempt, err := GradeQuizAttempt(assignment, *training.Quiz, map[string]string{
		strconv.FormatInt(q1.Id, 10): " B ",
		strconv.FormatInt(q2.Id, 10): "TRUE",
	})
	c.Assert(err, check.Equals, nil)
	c.Assert(attempt.Score, check.Equals, 100)
	c.Assert(attempt.Passed, check.Equals, true)

	updated, err := GetTrainingAssignmentByToken(assignment.Token)
	c.Assert(err, check.Equals, nil)
	c.Assert(updated.Status, check.Equals, AssignmentCompleted)
	c.Assert(updated.BestScore, check.Equals, 100)

	// A second assignment that gets only one of two right -> 50%, fails
	// (pass threshold is 80).
	assignment2, err := CreateTrainingAssignment(1, training.Id, nil, nil, "target2@example.com", "Target Two")
	c.Assert(err, check.Equals, nil)
	attempt2, err := GradeQuizAttempt(assignment2, *training.Quiz, map[string]string{
		strconv.FormatInt(q1.Id, 10): "wrong",
		strconv.FormatInt(q2.Id, 10): "true",
	})
	c.Assert(err, check.Equals, nil)
	c.Assert(attempt2.Score, check.Equals, 50)
	c.Assert(attempt2.Passed, check.Equals, false)
}

// TestGradeQuizAttemptMaxAttempts verifies the attempt limit is enforced
// and that reaching it without passing marks the assignment failed.
func (s *ModelsSuite) TestGradeQuizAttemptMaxAttempts(c *check.C) {
	training := s.createTraining(c, 80, 1)
	assignment, err := CreateTrainingAssignment(1, training.Id, nil, nil, "target3@example.com", "Target Three")
	c.Assert(err, check.Equals, nil)
	q1, q2 := training.Quiz.Questions[0], training.Quiz.Questions[1]

	_, err = GradeQuizAttempt(assignment, *training.Quiz, map[string]string{
		strconv.FormatInt(q1.Id, 10): "wrong",
		strconv.FormatInt(q2.Id, 10): "false",
	})
	c.Assert(err, check.Equals, nil)

	updated, err := GetTrainingAssignmentByToken(assignment.Token)
	c.Assert(err, check.Equals, nil)
	c.Assert(updated.Status, check.Equals, AssignmentFailed)

	_, err = GradeQuizAttempt(updated, *training.Quiz, map[string]string{
		strconv.FormatInt(q1.Id, 10): "b",
		strconv.FormatInt(q2.Id, 10): "true",
	})
	c.Assert(err, check.Not(check.Equals), nil)
}

// TestGetOrCreateTrainingAssignmentForResultDedups verifies a "both"
// trigger (click then submit) produces exactly one assignment per
// result, and the second call returns the same token as the first.
func (s *ModelsSuite) TestGetOrCreateTrainingAssignmentForResultDedups(c *check.C) {
	training := Training{Name: "Dedup Test"}
	c.Assert(PostTrainingForTenant(&training, 1, 1), check.Equals, nil)

	first, err := GetOrCreateTrainingAssignmentForResult(1, training.Id, 42, 7, "dup@example.com", "Dup")
	c.Assert(err, check.Equals, nil)
	c.Assert(first.Token, check.Not(check.Equals), "")

	second, err := GetOrCreateTrainingAssignmentForResult(1, training.Id, 42, 7, "dup@example.com", "Dup")
	c.Assert(err, check.Equals, nil)
	c.Assert(second.Id, check.Equals, first.Id)
	c.Assert(second.Token, check.Equals, first.Token)
}

