-- Add some new column to users table
ALTER TABLE users
    ADD COLUMN password character varying(255),
    ADD COLUMN email character varying(255),
    ADD COLUMN avatar character varying(255),
    ADD COLUMN nickname character varying(20),
    ADD COLUMN introduction character varying(255),
    ADD COLUMN status smallint DEFAULT 1,
    ADD COLUMN creator character varying(20);

-- Set the admin password, but other user has no password, they can't login to the system (headscale-panel).
UPDATE users SET password = '$2a$10$GkpxjdzwmPIIwdIXbQ91eOWJfQGK5J1M3da1/gmIpmfsEZIJ2O7kG' WHERE name = 'admin';
