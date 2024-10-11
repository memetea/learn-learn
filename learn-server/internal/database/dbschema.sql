-- Create question_banks table
CREATE TABLE question_banks (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL UNIQUE
);

-- Create question_types table
CREATE TABLE question_types (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL UNIQUE
);

-- Create questions table
CREATE TABLE questions (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    question_bank_id INTEGER NOT NULL,
    question_type_id INTEGER NOT NULL,
    content TEXT NOT NULL,
    explanation TEXT,
    FOREIGN KEY (question_bank_id) REFERENCES question_banks(id),
    FOREIGN KEY (question_type_id) REFERENCES question_types(id)
);

-- Create answer_options table
CREATE TABLE answer_options (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    question_id INTEGER NOT NULL,
    option_text TEXT NOT NULL,
    is_correct BOOLEAN NOT NULL,
    FOREIGN KEY (question_id) REFERENCES questions(id)
);

-- Create related_questions table
CREATE TABLE related_questions (
    question_id INTEGER,
    related_question_id INTEGER,
    PRIMARY KEY (question_id, related_question_id),
    FOREIGN KEY (question_id) REFERENCES questions(id),
    FOREIGN KEY (related_question_id) REFERENCES questions(id)
);

-- Insert some sample data
INSERT INTO question_banks (name) VALUES ('三年级英语'), ('四年级数学');
INSERT INTO question_types (name) VALUES ('选择题'), ('判断题');

-- Insert a multiple choice question
INSERT INTO questions (question_bank_id, question_type_id, content, explanation)
VALUES (1, 1, 'What is the capital of France?', 'Paris is the capital and largest city of France.');

-- Insert answer options for the multiple choice question
INSERT INTO answer_options (question_id, option_text, is_correct) VALUES 
(1, 'London', 0),
(1, 'Berlin', 0),
(1, 'Paris', 1),
(1, 'Madrid', 0);

-- Insert a true/false question
INSERT INTO questions (question_bank_id, question_type_id, content, explanation)
VALUES (1, 2, 'Is the Earth flat?', 'The Earth is approximately spherical in shape.');

-- Insert answer options for the true/false question
INSERT INTO answer_options (question_id, option_text, is_correct) VALUES 
(2, 'True', 0),
(2, 'False', 1);

INSERT INTO related_questions (question_id, related_question_id) VALUES (1, 2);