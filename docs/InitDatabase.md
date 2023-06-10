# Initializing the database

> Before executing the database operation scripts, make sure you have installed the `psql` command in your execution environment.

To use the scripts, you need to use the `database.sh` file in the project's root directory and the `.sql` files.
The command format is as follows. Please choose the appropriate SQL file based on your needs:
```shell
./database.sh -h <host or ip> -P <port> -u <username> -p <password> -d <database> -f <sql file>
```

- `edit_users_table.sql`: Directly modifies the original `users` table.
- `backup_copy_new_users.sql`: First backs up the `users` table as `src_users`, then rebuilds the `users` table using the existing data.
- `set_admin_password.sql`: Resets the password for the `admin` user to `123456`.
- `set_password.sql`: Resets the password for all users to `123456`.