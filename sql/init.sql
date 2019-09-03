-- To create and set up the database, run the Database Creation code after setting the following fields:
-- * DATABASE_NAME : The name of the database.  Eg: nate_mlb
-- * DATABASE_IP_ADDRESS : The ip address of the database.  Eg: 127.0.0.1 if running on localhost
-- * DATABASE_PORT : The port used to connect to the database.  Eg: 5432
-- * DATABASE_USERNAME: The username of the user not run all database operations as.  Eg: nate
-- * DATABASE_PASSWORD: The password of the user.
-- The DATABASE_URL environment variable will be postgres://<DATABASE_USERNAME>:<DATABASE_PASSWORD>@<DATABASE_IP_ADDRESS>:<DATABASE_PORT>/<DATABASE_NAME>

-- Database Creation code:
-- $ sudo -u postgres psql
-- CREATE DATABASE <DATABASE_NAME>;
-- CREATE USER <DATABASE_USERNAME> WITH ENCRYPTED PASSWORD '<DATABASE_PASSWORD>';
-- GRANT ALL PRIVILEGES ON DATABASE <DATABASE_NAME> TO <DATABASE_USERNAME>;
-- \quit

-- Initialization:
-- psql -d <DATABASE_NAME> -U <DATABASE_USERNAME> -h <DATABASE_IP_ADDRESS> -W
-- [enter <DATABASE_PASSWORD>]
-- -- (run sql code in this file)
-- \quit

-- Initialize Admin Password:
-- $ printf "UPDATE users SET password='%s' WHERE username='admin';" \
--   $(curl -s http://<SERVER_HOST>:<SERVER_PORT>/admin/password?v=<ADMIN_PASSWORD>) |
--   psql -d <DATABASE_NAME> -U <DATABASE_USER> -h <DATABASE_IP_ADDRESS> -W
-- [enter <DATABASE_PASSWORD>]
-- curl -X POST -d 'password=<ADMIN_PASSWORD>' http://<SERVER_HOST:<SERVER_PORT>/admin/password
-- Note that the admin password does not have to be the database password (it should be different)

-- Remove the database and the user
-- $ sudo -u postgres psql
-- DROP DATABASE <DATABASE_NAME>;
-- DROP USER <DATABASE_USERNAME>;
-- \quit

-- DROP TABLE IF EXISTS users;
