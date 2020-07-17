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
                      params={"email": USER1_EMAIL, "verification": code, "tz": "Asia/Shanghai", "lang": "en-US"}, data={"password": USER1_PWD})
    j = r.json()
    assert j["code"] == 0


@test_case
def login_user(c: Context) -> None:
    r = requests.post(c.service_url + "v1/login",
                      params={"email": USER1_EMAIL}, data={"password": USER1_PWD})
    j = r.json()
    print(j)
    assert j["code"] == 0
    global USER1_TOKEN
    USER1_TOKEN = j["data"]["token"]


@test_case
def user_info(c: Context) -> None:
    r = requests.post(c.service_url + "v1/user/info",
                      headers={"WS-User-Token": USER1_TOKEN})
    j = r.json()
    print(j)
    assert j["code"] == 0
    assert j["data"]["language"] == "en-US"
    assert j["data"]["timeZone"] == "Asia/Shanghai"


@test_case
def update_user_setting(c: Context) -> None:
    # Note: one can also update multiple value at the same time
    r = requests.post(c.service_url + "v1/user/update", headers={"WS-User-Token": USER1_TOKEN},
                      params={"tz": "Australia/Melbourne"})
    j = r.json()
    print(j)
    assert j["code"] == 0

    r = requests.post(c.service_url + "v1/user/update", headers={"WS-User-Token": USER1_TOKEN},
                      params={"lang": "zh-Hans"})
    j = r.json()
    print(j)
    assert j["code"] == 0

    r = requests.post(c.service_url + "v1/user/info",
                      headers={"WS-User-Token": USER1_TOKEN})
    j = r.json()
    print(j)
    assert j["code"] == 0
    assert j["data"]["language"] == "zh-Hans"
    assert j["data"]["timeZone"] == "Australia/Melbourne"
