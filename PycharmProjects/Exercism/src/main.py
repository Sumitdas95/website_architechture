import sqlite3


def user_db_creation(first_name, last_name, email, password):
    conn = sqlite3.connect('data_store/login.db')
    cursor = conn.cursor()

    # Create table if not exists
    cursor.execute('''
        CREATE TABLE IF NOT EXISTS users (
            user_id INTEGER PRIMARY KEY,
            first_name TEXT NOT NULL,
            last_name TEXT NOT NULL,
            email TEXT NOT NULL UNIQUE,
            password TEXT NOT NULL
        )
    ''')
    cursor.execute('SELECT * FROM users WHERE email = ?', (email,))
    existing_user = cursor.fetchone()
    if existing_user:
        print("Email address already exists")
    else:
        # Insert user details
        cursor.execute('''INSERT INTO users (first_name, last_name, email, password) 
                                                              VALUES (?, ?, ?, ?)''',
                       (first_name, last_name, email, password))

    conn.commit()
    conn.close()


def main():
    user_db_creation("sumit", "das", "sumit@gmail.com", "sumit@123")
    user_db_creation("suvajit", "majumder", "suvajit@gmail.com", "suvajit@123")
    user_db_creation("sudeshna", "majumder", "sudeshna@gmail.com", "sudeshna@123")
    user_db_creation("piyali", "chowdhury", "piyali@gmail.com", "piyali@123")


if __name__ == "__main__":
    main()