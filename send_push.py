#!/bin/python3
import hmac
from hashlib import sha256
from http import HTTPStatus
from os import getenv

from requests import post


def main():
    secret = getenv("WEBHOOK_SECRET", "")
    content = ""
    hash_object = hmac.new(
        secret.encode("utf-8"), msg=content.encode("utf-8"), digestmod=sha256
    )
    req = post(
        "https://unspaced-aimee-flauntingly.ngrok-free.dev",
        headers={
            "x-github-event": "push",
            "X-Hub-Signature-256": f"sha256={hash_object.hexdigest()}",
        },
    )

    if req.status_code == HTTPStatus.OK:
        print("push sent succesfully")
        exit(0)
    else:
        print(f"push failed with status {req.status_code}")
        exit(1)


if __name__ == "__main__":
    main()
