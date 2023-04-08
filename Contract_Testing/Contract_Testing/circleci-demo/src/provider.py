#!/usr/env/bin python
import datetime
import json
import uuid

from flask import Flask, abort, jsonify, request

fakedb = {}

app = Flask(__name__)


@app.route('/_pact/provider_states', methods=['POST'])
def provider_states():
    print('json..', request.json)
    return jsonify({'result': request.json['state']})


def setup_admin():
    pass


@app.route('/is_superuser/<username>')
def is_super_user(username):
    print('in is_super_user function')
    response = jsonify(
        {
            'status': 'failed',
        }
    )
    response.status_code = 400
    if username != 'admin':
        return response
    response = jsonify(
        {
            'status': 'success',
        }
    )
    return response


if __name__ == '__main__':
    app.run(debug=True, port=5001)
