from flask import Flask, request, jsonify, make_response
import sqlite3
import uuid


app = Flask(__name__)


def session_db_creation(user_id, session_id):
    conn = sqlite3.connect('data_store/session.db')
    cursor = conn.cursor()

    # Create table if not exists
    cursor.execute('''CREATE TABLE IF NOT EXISTS session 
                          (user_id INTEGER PRIMARY KEY, 
                          session_id TEXT,
                          FOREIGN KEY (user_id) REFERENCES users(user_id))''')

    cursor.execute('SELECT * FROM session WHERE user_id = ? AND session_id = ?', (user_id, session_id))
    if cursor.fetchone():
        print("wrong input: session already exists for this user_id and session_id")

    else:
        # Insert session details
        cursor.execute('''INSERT INTO session (user_id, session_id) 
                                          VALUES (?, ?)''', (user_id, session_id))
    conn.commit()
    conn.close()


def check_user_credentials(email, password):
    conn = sqlite3.connect('data_store/login.db')
    cursor = conn.cursor()

    cursor.execute('SELECT * FROM users WHERE email = ? AND password = ?', (email, password))
    user_data = cursor.fetchone()
    if user_data:
        return user_data
    return False


@app.route('/login', methods=['POST'])
def login():
    data = request.json
    email = data.get('email')
    password = data.get('password')
    user = check_user_credentials(email, password)
    if user:
        session_id = str(uuid.uuid4())
        session_db_creation(user[0], session_id)
        response = make_response(jsonify({
            "message": "Login successful"
        }))
        response.status_code = 200
        response.headers['set-cookie'] = session_id
        return response
    else:
        return jsonify({"message": "Invalid email or password"}), 401


@app.route('/home', methods=['GET'])
def home():
    session_id = request.headers.get('Cookie')
    if not session_id:
        return jsonify({"message": "Cookie is missing"}), 400
    conn = sqlite3.connect('data_store/session.db')
    cursor = conn.cursor()
    cursor.execute('SELECT user_id FROM session WHERE session_id = ?', (session_id,))
    session_data = cursor.fetchone()

    if session_data:
        user_id = session_data[0]
        conn_user = sqlite3.connect('data_store/login.db')
        cursor_user = conn_user.cursor()
        cursor_user.execute('SELECT first_name FROM users WHERE user_id = ?', (user_id,))
        user_data = cursor_user.fetchone()
        if user_data:
            return jsonify({"message": f"Welcome to Home {user_data[0]} !!"})

    return jsonify({"message": "Invalid session id"}), 401


def main():
    app.run(debug=True)


if __name__ == "__main__":
    main()