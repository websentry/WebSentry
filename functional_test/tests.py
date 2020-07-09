from typing import *
from helper import DBHelper
import requests

test_cases = []


def test_case(func):
    # register test case
    test_cases.append((func.__name__, func))
    return func


class Context:
    def __init__(self, db_helper: DBHelper, service_url: str) -> None:
        self.db_helper = db_helper
        self.service_url = service_url


USER1_EMAIL = "test@test.com"
USER1_PWD = "pwd123456"
USER1_TOKEN = None


@test_case
def register_user(c: Context) -> None:
    # verification code
    r = requests.post(c.service_url + "v1/get_verification",
                      params={"email": USER1_EMAIL})
    j = r.json()
    assert j["code"] == 0
    code = c.db_helper.get_email_verification_code(USER1_EMAIL)
    print("verification code: {}".format(code))

    # register
    r = requests.post(c.service_url + "v1/create_user",
                      params={"email": USER1_EMAIL, "verification": code}, data={"password": USER1_PWD})
    j = r.json()
    assert j["code"] == 0
