"""pact test for user service client"""

import json
import logging
import os
import sys

import pytest
import requests
from requests.auth import HTTPBasicAuth

from src.consumer import UserClient
from pact import Consumer, Like, Provider, Term

log = logging.getLogger(__name__)
logging.basicConfig(level=logging.INFO)

PACT_UPLOAD_URL = (
    "http://127.0.0.1/pacts/provider/UserService/consumer"
    "/UserServiceClient/version"
)
PACT_FILE = "userserviceclient-userservice.json"
PACT_BROKER_USERNAME = "pactbroker"
PACT_BROKER_PASSWORD = "pactbroker"

PACT_MOCK_HOST = '127.0.0.1'
PACT_MOCK_PORT = 5001
PACT_DIR = os.path.dirname(os.path.realpath(__file__))


@pytest.fixture
def client():
    return UserClient(
        'http://{host}:{port}'
        .format(host=PACT_MOCK_HOST, port=PACT_MOCK_PORT)
    )


def push_to_broker(version):
    """TODO: see if we can dynamically learn the pact file name, version, etc."""
    with open(os.path.join(PACT_DIR, PACT_FILE), 'rb') as pact_file:
        pact_file_json = json.load(pact_file)

    basic_auth = HTTPBasicAuth(PACT_BROKER_USERNAME, PACT_BROKER_PASSWORD)

    log.info("Uploading pact file to pact broker...")

    r = requests.put(
        "{}/{}".format(PACT_UPLOAD_URL, version),
        auth=basic_auth,
        json=pact_file_json
    )
    if not r.ok:
        log.error("Error uploading: %s", r.content)
        r.raise_for_status()


@pytest.fixture(scope='session')
def pact(request):
    pact = Consumer('UserServiceClient').has_pact_with(
        Provider('UserService'), host_name=PACT_MOCK_HOST, port=PACT_MOCK_PORT,
        pact_dir=PACT_DIR)
    pact.start_service()
    yield pact
    pact.stop_service()


def test_is_super_user_with_valid_admin_credentials(pact, client):
    expected = {
        'status': 'success',
    }

    (pact
     .given('An admin should logged in')
     .upon_receiving('A request for admin')
     .with_request('get', '/is_superuser/admin')
     .will_respond_with(200, body=expected))

    with pact:
        result = client.is_super_user('admin')
        assert result == expected
