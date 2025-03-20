CREATE DATABASE IF NOT EXISTS muzz;

USE muzz;

CREATE TABLE IF NOT EXISTS decisions (
    id INT AUTO_INCREMENT PRIMARY KEY,
    actor_user_id VARCHAR(36) NOT NULL,
    recipient_user_id VARCHAR(36) NOT NULL,
    liked BOOLEAN NOT NULL,
    unix_timestamp BIGINT DEFAULT (UNIX_TIMESTAMP()),
    UNIQUE KEY unique_decision (actor_user_id, recipient_user_id),
    INDEX (recipient_user_id),
    INDEX (actor_user_id)
);
