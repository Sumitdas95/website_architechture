from flask import Flask, request, make_response, jsonify
import datetime
import jwt
import sqlite3


app = Flask(__name__)

SECRET_KEY = 'secret-key'
substring = '.com'


def check_user_credentials(email, password):
    conn = sqlite3.connect('data_store/login.db')
    cursor = conn.cursor()

    cursor.execute('SELECT * FROM users WHERE email = ? AND password = ?', (email, password))
    user_data = cursor.fetchone()
    if user_data:
        return user_data
    return False


def get_user_detail(email):
    conn_user = sqlite3.connect('data_store/login.db')
    cursor_user = conn_user.cursor()
    cursor_user.execute('SELECT first_name FROM users WHERE email = ?', (email,))
    user_data = cursor_user.fetchone()
    if user_data:
        return user_data[0]
    else:
        return None


def generate_jwt(email):
    payload = {
        'email': email,
        'exp': datetime.datetime.utcnow() + datetime.timedelta(minutes=2)
    }
    token = jwt.encode(payload, SECRET_KEY, algorithm='HS256')
    return token


def decode_jwt(token):
    try:
        decoded_payload = jwt.decode(token, SECRET_KEY, algorithms=['HS256'])
        email = decoded_payload['email']
        return email
    except jwt.ExpiredSignatureError:
        return 'Token has expired!'
    except jwt.InvalidTokenError:
        return 'Invalid token!'


@app.route('/login', methods=['POST'])
def login():
    data = request.json
    email = data.get('email')
    password = data.get('password')
    user = check_user_credentials(email, password)
    if user:
        token = generate_jwt(email)
        response = make_response(jsonify({
            "message": "Login successful"
        }))
        response.status_code = 200
        response.headers['Authorization'] = f'Bearer {token}'
        return response
    else:
        return jsonify({"message": "Invalid email or password"}), 401


@app.route('/home', methods=['GET'])
def home():
    token = request.headers.get('Authorization').split(" ")[1]
    data = decode_jwt(token)

    if substring in data:
        username = get_user_detail(data)
        if username:
            return jsonify({"message": f"Welcome to home {username} !!"})
        else:
            return jsonify({"message": "Invalid email id"})

    return jsonify({"message": data}), 401


def main():
    app.run(debug=True)


if __name__ == "__main__":
    main()