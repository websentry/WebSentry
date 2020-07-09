from typing import *
import sys
import argparse
import os
import tempfile
import time
import subprocess
import json
import requests
import traceback
from helper import DBHelper
from tests import Context, test_cases


class Bcolors:
    HEADER = '\033[95m'
    OKBLUE = '\033[94m'
    OKGREEN = '\033[92m'
    WARNING = '\033[93m'
    FAIL = '\033[91m'
    ENDC = '\033[0m'
    BOLD = '\033[1m'
    UNDERLINE = '\033[4m'


def main() -> None:
    # parse arguments
    parser = argparse.ArgumentParser(
        description='Functional test')
    parser.add_argument('--db', type=str,
                        help='db connection string')
    parser.add_argument('--skip_clean_db', type=bool,
                        help='db will be cleaned by default')
    parser.add_argument('--port', type=int,
                        help='port for websentry service')
    if len(sys.argv) == 1:
        parser.print_help(sys.stderr)
        sys.exit(1)

    args = parser.parse_args()
    print(Bcolors.OKBLUE, args, Bcolors.ENDC)

    start = time.time()
    ok = run_test(args.db, args.skip_clean_db, args.port)
    print()
    print("=" * 40)
    print("Time used: {:.2f}s".format(time.time() - start))
    if ok:
        print(Bcolors.OKGREEN + "PASS" + Bcolors.ENDC)
    else:
        print(Bcolors.FAIL + "FAIL" + Bcolors.ENDC)
        sys.exit(1)


def run_test(db: str, skip_clean_db: bool, port: int) -> bool:
    root_path = os.path.dirname(os.path.dirname(os.path.realpath(__file__)))
    os.chdir(root_path)
    tmp_dir = tempfile.TemporaryDirectory()
    is_ok = True
    service_p = None
    db_helper = None
    service_url = "http://127.0.0.1:{}/".format(port)

    try:
        db_helper = DBHelper(db)
        build_config(tmp_dir.name, db, port)
        if not skip_clean_db:
            db_helper.drop_all_tables()
        service_p = start_service(tmp_dir.name)
        wait_for_service(service_p, service_url)
        print()

        # start running tests
        c = Context(db_helper, service_url)
        for i, (name, f) in enumerate(test_cases):
            print("({}/{}) {}:".format(i+1, len(test_cases), name))
            try:
                f(c)
                print(Bcolors.OKGREEN+"OK"+Bcolors.ENDC)
                print()
            except Exception as e:
                print(Bcolors.FAIL+"FAIL"+Bcolors.ENDC)
                print(e)
                traceback.print_exc()
                print()
                is_ok = False
                break

    except Exception as e:
        print_service_log(service_p)
        print(Bcolors.FAIL + "Exception occurred during the test:" + Bcolors.ENDC)
        print(e)
        traceback.print_exc()
        return False
    finally:
        tmp_dir.cleanup()
        if service_p is not None:
            service_p.terminate()
        if db_helper is not None:
            db_helper.close()

    if is_ok == False:
        print_service_log(service_p)
    return is_ok


def build_config(tmp_path: str, db: str, port: int) -> None:
    config = {
        "releaseMode": False,
        "addr": "127.0.0.1:{}".format(port),
        "database": {
            "dataSourceName": db,
            "type": "postgres"
        },
        "verificationEmail": {
            "server": "your_smtp_server",
            "port": 587,
            "email": "example@example.com",
            "password": "password"
        },
        "fileStoragePath": tmp_path,
        "slaveKey": "testkey",
        "tokenSecretKey": "secretkey",
        "backendUrl": "http://127.0.0.1:{}/".format(port),
        "crosAllowOrigins": ["*"],
        "forwardedByClientIP": True
    }
    with open(os.path.join(tmp_path, "config.json"), 'w') as outfile:
        json.dump(config, outfile)


def start_service(tmp_path: str) -> subprocess.Popen:
    print("starting websentry service")
    service_p = subprocess.Popen(
        ["./websentry", "-c", os.path.join(tmp_path, "config.json")], stdout=subprocess.PIPE, stderr=subprocess.STDOUT)
    return service_p


def wait_for_service(service_p: subprocess.Popen, service_url: str) -> None:
    for i in range(4):
        check_service_running(service_p)
        try:
            r = requests.get(service_url + "ping")
            if r.text == "pong":
                return
        except Exception as e:
            pass
        time.sleep(0.5)
    r = requests.get(service_url + "ping")
    assert r.text == "pong"


def print_service_log(service_p: subprocess.Popen) -> None:
    print()
    print(Bcolors.OKBLUE + "WebSentry log:" + Bcolors.ENDC)
    if service_p is not None:
        output = service_p.stdout.read().decode("utf-8")
        print(output)
    print()


def check_service_running(service_p: subprocess.Popen) -> None:
    if service_p.poll() is not None:
        raise RuntimeError("WebSentry service exit unexpectedly")


if __name__ == "__main__":
    main()
