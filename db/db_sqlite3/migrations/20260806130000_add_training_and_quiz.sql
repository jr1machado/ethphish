-- +goose Up
CREATE TABLE IF NOT EXISTS trainings (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    tenant_id INTEGER NOT NULL DEFAULT 1,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    created_by INTEGER NOT NULL,
    created_at DATETIME,
    updated_at DATETIME
);

CREATE TABLE IF NOT EXISTS training_lessons (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    training_id INTEGER NOT NULL,
    position INTEGER NOT NULL DEFAULT 0,
    title VARCHAR(255) NOT NULL,
    html TEXT
);

CREATE TABLE IF NOT EXISTS training_quizzes (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    training_id INTEGER NOT NULL,
    pass_percent INTEGER NOT NULL DEFAULT 80,
    max_attempts INTEGER NOT NULL DEFAULT 0
);

CREATE TABLE IF NOT EXISTS quiz_questions (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    quiz_id INTEGER NOT NULL,
    position INTEGER NOT NULL DEFAULT 0,
    type VARCHAR(30) NOT NULL DEFAULT 'multiple_choice',
    text TEXT NOT NULL,
    options TEXT,
    correct_answer VARCHAR(255) NOT NULL
);

CREATE TABLE IF NOT EXISTS training_assignments (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    tenant_id INTEGER NOT NULL DEFAULT 1,
    training_id INTEGER NOT NULL,
    campaign_id INTEGER,
    result_id INTEGER,
    email VARCHAR(255) NOT NULL,
    name VARCHAR(255),
    token VARCHAR(255) NOT NULL,
    status VARCHAR(30) NOT NULL DEFAULT 'assigned',
    attempts INTEGER NOT NULL DEFAULT 0,
    best_score INTEGER NOT NULL DEFAULT 0,
    created_at DATETIME,
    completed_at DATETIME
);

CREATE TABLE IF NOT EXISTS training_lesson_views (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    assignment_id INTEGER NOT NULL,
    lesson_id INTEGER NOT NULL,
    viewed_at DATETIME
);

CREATE TABLE IF NOT EXISTS quiz_attempts (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    assignment_id INTEGER NOT NULL,
    attempt_number INTEGER NOT NULL,
    score INTEGER NOT NULL,
    passed BOOLEAN NOT NULL DEFAULT 0,
    answers TEXT,
    created_at DATETIME
);

ALTER TABLE campaigns ADD COLUMN training_id INTEGER;
ALTER TABLE campaigns ADD COLUMN training_trigger VARCHAR(20) NOT NULL DEFAULT '';

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
