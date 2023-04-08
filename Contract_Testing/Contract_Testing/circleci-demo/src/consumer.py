import requests


class UserClient(object):
    def __init__(self, base_uri):
        self.base_uri = base_uri

    def is_super_user(self, user_name=None):
        url = f'{self.base_uri}/is_superuser/{user_name}'
        response = requests.get(url)
        print(response.json())
        if response.status_code == 400:
            return None
        return response.json()
