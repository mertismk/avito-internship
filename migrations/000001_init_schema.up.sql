-- Создаем таблицу команд
CREATE TABLE IF NOT EXISTS teams (
                                     team_name VARCHAR(255) PRIMARY KEY,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
    );

-- Создаем таблицу пользователей
CREATE TABLE IF NOT EXISTS users (
                                     user_id VARCHAR(255) PRIMARY KEY,
    username VARCHAR(255) NOT NULL,
    team_name VARCHAR(255) NOT NULL REFERENCES teams(team_name) ON DELETE CASCADE,
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
    );

-- Создаем индексы для быстрого поиска пользователей
CREATE INDEX idx_users_team_name ON users(team_name);
CREATE INDEX idx_users_is_active ON users(is_active);

-- Создаем таблицу Pull Request'ов
CREATE TABLE IF NOT EXISTS pull_requests (
                                             pull_request_id VARCHAR(255) PRIMARY KEY,
    pull_request_name VARCHAR(255) NOT NULL,
    author_id VARCHAR(255) NOT NULL REFERENCES users(user_id),
    status VARCHAR(10) NOT NULL CHECK (status IN ('OPEN', 'MERGED')),
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    merged_at TIMESTAMP
    );

-- Создаем индексы для фильтрации по автору и статусу
CREATE INDEX idx_pull_requests_author_id ON pull_requests(author_id);
CREATE INDEX idx_pull_requests_status ON pull_requests(status);

-- Создаем таблицу ревьюверов на PR
CREATE TABLE IF NOT EXISTS pr_reviewers (
                                            pull_request_id VARCHAR(255) NOT NULL REFERENCES pull_requests(pull_request_id) ON DELETE CASCADE,
    user_id VARCHAR(255) NOT NULL REFERENCES users(user_id),
    assigned_at TIMESTAMP NOT NULL DEFAULT NOW(),
    PRIMARY KEY (pull_request_id, user_id)
    );

-- Создаем индекс для быстрого поиска PR по ревьюверу
CREATE INDEX idx_pr_reviewers_user_id ON pr_reviewers(user_id);