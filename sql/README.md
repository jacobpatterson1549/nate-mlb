## Database Setup
The nate-mlb server expects a PostgreSQL database to be supplied.

### Create database
The bash template below can be edited and run to create the database and user.  It outputs the DATABASE_URL (data source name) if the database and user are successfully created.
```bash
PGDATABASE="nate_mlb_db" \
PGUSER="nate" \
PGPASSWORD="nate12345" \
PGHOSTADDR="127.0.0.1" \
PGPORT="5432" \
sh -c ' \
sudo -u postgres psql \
-c "CREATE DATABASE $PGDATABASE" \
-c "CREATE USER $PGUSER WITH ENCRYPTED PASSWORD '"'"'$PGPASSWORD'"'"'" \
-c "GRANT ALL PRIVILEGES ON DATABASE $PGDATABASE TO $PGUSER" \
&& echo DATABASE_URL=postgres://$PGUSER:$PGPASSWORD@$PGHOSTADDR:$PGPORT/$PGDATABASE \
'
```

### Remove Database
The bash template below can be edited and run to **permanently** delete the database data and user.  This includes saved data, tables, indexes, functions, and user credentials.
```bash
PGDATABASE="nate_mlb_db" \
PGUSER="nate" \
sh -c '\
sudo -u postgres psql \
-c "DROP DATABASE $PGDATABASE" \
-c "DROP USER $PGUSER" \
'
```
