-- +goose Up
CREATE TABLE IF NOT EXISTS trainings (
    id integer primary key auto_increment,
    tenant_id integer NOT NULL DEFAULT 1,
    name varchar(255) NOT NULL,
    description text,
    created_by integer NOT NULL,
    created_at datetime,
    updated_at datetime
);

CREATE TABLE IF NOT EXISTS training_lessons (
    id integer primary key auto_increment,
    training_id integer NOT NULL,
    position integer NOT NULL DEFAULT 0,
    title varchar(255) NOT NULL,
    html text
);

CREATE TABLE IF NOT EXISTS training_quizzes (
    id integer primary key auto_increment,
    training_id integer NOT NULL,
    pass_percent integer NOT NULL DEFAULT 80,
    max_attempts integer NOT NULL DEFAULT 0
);

CREATE TABLE IF NOT EXISTS quiz_questions (
    id integer primary key auto_increment,
    quiz_id integer NOT NULL,
    position integer NOT NULL DEFAULT 0,
    type varchar(30) NOT NULL DEFAULT 'multiple_choice',
    text text NOT NULL,
    options text,
    correct_answer varchar(255) NOT NULL
);

CREATE TABLE IF NOT EXISTS training_assignments (
    id integer primary key auto_increment,
    tenant_id integer NOT NULL DEFAULT 1,
    training_id integer NOT NULL,
    campaign_id integer,
    result_id integer,
    email varchar(255) NOT NULL,
    name varchar(255),
    token varchar(255) NOT NULL,
    status varchar(30) NOT NULL DEFAULT 'assigned',
    attempts integer NOT NULL DEFAULT 0,
    best_score integer NOT NULL DEFAULT 0,
    created_at datetime,
    completed_at datetime
);

CREATE TABLE IF NOT EXISTS training_lesson_views (
    id integer primary key auto_increment,
    assignment_id integer NOT NULL,
    lesson_id integer NOT NULL,
    viewed_at datetime
);

CREATE TABLE IF NOT EXISTS quiz_attempts (
    id integer primary key auto_increment,
    assignment_id integer NOT NULL,
    attempt_number integer NOT NULL,
    score integer NOT NULL,
    passed boolean NOT NULL DEFAULT false,
    answers text,
    created_at datetime
);

ALTER TABLE campaigns ADD COLUMN training_id integer;
ALTER TABLE campaigns ADD COLUMN training_trigger varchar(20) NOT NULL DEFAULT '';

CREATE INDEX idx_trainings_tenant_id ON trainings(tenant_id);
CREATE INDEX idx_training_lessons_training_id ON training_lessons(training_id);
CREATE INDEX idx_training_quizzes_training_id ON training_quizzes(training_id);
CREATE INDEX idx_quiz_questions_quiz_id ON quiz_questions(quiz_id);
CREATE INDEX idx_training_assignments_tenant_id ON training_assignments(tenant_id);
CREATE INDEX idx_training_assignments_token ON training_assignments(token);
CREATE INDEX idx_training_assignments_campaign_result ON training_assignments(campaign_id, result_id);
CREATE INDEX idx_training_lesson_views_assignment_id ON training_lesson_views(assignment_id);
CREATE INDEX idx_quiz_attempts_assignment_id ON quiz_attempts(assignment_id);

-- +goose Down
ALTER TABLE campaigns DROP COLUMN training_id;
ALTER TABLE campaigns DROP COLUMN training_trigger;
DROP TABLE quiz_attempts;
DROP TABLE training_lesson_views;
DROP TABLE training_assignments;
DROP TABLE quiz_questions;
DROP TABLE training_quizzes;
DROP TABLE training_lessons;
DROP TABLE trainings;
