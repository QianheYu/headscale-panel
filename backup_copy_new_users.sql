-- Rename the users table to src_users
ALTER TABLE users RENAME TO src_users;

-- Create the new users table
CREATE TABLE users (
                       id bigint NOT NULL DEFAULT nextval('users_id_seq'::regclass),
                       created_at timestamp with time zone,
                       updated_at timestamp with time zone,
                       deleted_at timestamp with time zone,
                       name text unique,
                       password character varying(255),
                       email character varying(255),
                       avatar character varying(255),
                       nickname character varying(20),
                       introduction character varying(255),
                       status smallint DEFAULT 1,
                       creator character varying(20)
);

-- Rename admin to avoid confict if admin has already exist
UPDATE src_users SET name = 'admin2' WHERE name = 'admin' ;

-- Copy the data from src_users to users, and set the all user password
INSERT INTO users (id, created_at, updated_at, deleted_at, name, creator)
SELECT id, created_at, updated_at, deleted_at, name, 'admin'
FROM src_users;

-- Drop the src_users table
DROP TABLE src_users;
