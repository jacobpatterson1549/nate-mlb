## Database Setup
The nate-mlb server expects a PostgreSQL database to be supplied.

### Create database
To create and set up the database, run the Database Creation code after setting the following fields
* DATABASE_NAME : The name of the database.  Eg: nate_mlb
* DATABASE_IP_ADDRESS : The ip address of the database.  Eg: 127.0.0.1 if running on localhost
* DATABASE_PORT : The port used to connect to the database.  Eg: 5432
* DATABASE_USERNAME: The username of the user not run all database operations as.  Eg: nate
* DATABASE_PASSWORD: The password of the user.

The DATABASE_URL environment variable will be: `postgres://<DATABASE_USERNAME>:<DATABASE_PASSWORD>@<DATABASE_IP_ADDRESS>:<DATABASE_PORT>/<DATABASE_NAME>`

The following template can be used to create the database and user.
```bash
# start psql as postgres:
sudo -u postgres psql
```
```sql
-- in psql console as postgres:
CREATE DATABASE <DATABASE_NAME>;
CREATE USER <DATABASE_USERNAME> WITH ENCRYPTED PASSWORD '<DATABASE_PASSWORD>';
GRANT ALL PRIVILEGES ON DATABASE <DATABASE_NAME> TO <DATABASE_USERNAME>;
\quit
```

### Remove Database
To remove all remnants of the database, run the code below.  This permanently deletes the database, which includes saved data, tables, indexes, functions, and user credentials.
```sql
-- as postgres
DROP DATABASE <DATABASE_NAME>;
DROP USER <DATABASE_USERNAME>;
```
