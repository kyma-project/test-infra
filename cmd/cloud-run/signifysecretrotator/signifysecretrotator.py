"""Simple PubSub handler to rotate signify certificates"""

import datetime
import os
import base64
import json
import sys
import traceback
from typing import Any, Dict, List
from flask import Flask, Response, request, make_response
from cryptography import x509
from cryptography.hazmat.primitives.asymmetric import rsa
from cryptography.hazmat.primitives import serialization
from cryptography.hazmat.primitives.serialization import Encoding, PrivateFormat
from requests import HTTPError
from secretmanager import client
from pylogger.logger import Logger
from signify.client import SignifyClient

app = Flask(__name__)
project_id: str = os.getenv("PROJECT_ID", "sap-kyma-prow")
component_name: str = os.getenv("COMPONENT_NAME", "signify-certificate-rotator")
application_name: str = os.getenv("APPLICATION_NAME", "secret-rotator")
secret_rotate_message_type = os.getenv("SECRET_ROTATE_MESSAGE_TYPE", "signify")
rsa_key_size: int = 4096


@app.route("/", methods=["POST"])
def rotate_signify_secret() -> Response:
    """HTTP webhook handler for rotating Signify secrets."""
    logger = Logger(
        component_name=component_name,
        application_name=application_name,
        request=request,
    )

    try:
        sm_client = client.SecretManagerClient()

        pubsub_message: Dict[str, Any] = get_pubsub_message()

        secret_rotate_msg: Dict[str, Any] = extract_message_data(pubsub_message)

        validate_message(secret_rotate_msg)

        secret_id: str = secret_rotate_msg["name"]
        secret_data: Dict[str, Any] = sm_client.get_secret(secret_id)

        signify_client = SignifyClient(
            token_url=secret_data["tokenURL"],
            certificate_service_url=secret_data["certServiceURL"],
            client_id=secret_data["clientID"],
        )

        old_cert_data: bytes = base64.b64decode(secret_data["certData"])
        old_pk_data: bytes = base64.b64decode(secret_data["privateKeyData"])

        if "password" in secret_data and secret_data["password"] != "":
            old_pk_data = decrypt_private_key(
                old_pk_data, secret_data["password"].encode()
            )

        new_private_key: rsa.RSAPrivateKey = rsa.generate_private_key(
            # Public exponent is standarised as 65537
            # see: https://cryptography.io/en/latest/hazmat/primitives/asymmetric/rsa/#cryptography.hazmat.primitives.asymmetric.rsa.generate_private_key
            public_exponent=65537,
            key_size=rsa_key_size,
        )

        access_token: str = signify_client.fetch_access_token(
            certificate=old_cert_data,
            private_key=old_pk_data,
        )

        created_at: str = datetime.datetime.now().strftime("%d-%m-%Y %H:%M:%S")

        new_certs: List[x509.Certificate] = signify_client.fetch_new_certificate(
            cert_data=old_cert_data,
            private_key=new_private_key,
            access_token=access_token,
        )

        new_secret_data: Dict[str, Any] = prepare_new_secret(
            new_certs, new_private_key, secret_data, created_at
        )

        sm_client.set_secret(secret_id, json.dumps(new_secret_data))

        logger.log_info(f"Certificate rotated successfully at {created_at}")

        return "Certificate rotated successfully"
    except (HTTPError, ValueError, TypeError) as exc:
        return prepare_error_response(exc, logger)


def validate_message(message: dict[str, Any]) -> None:
    """Raises error when received message struct is invalid"""

    # Pub/Sub topic handle multiple secret rotator components
    # verify if we should handle that message
    if message.get("labels", {}).get("type") != secret_rotate_message_type:
        raise TypeError("Invalid or unknown type value")


def prepare_new_secret(
    certificates: List[x509.Certificate],
    private_key: rsa.RSAPrivateKey,
    secret_data: Dict[str, Any],
    created_at: str,
) -> Dict[str, Any]:
    """Prepares new secret data with updated certificates and private key."""

    # format certificates
    certs_string: str = ""

    for cert in certificates:
        certs_string += f"subject={cert.subject.rfc4514_string()}\n"
        certs_string += f"issuer={cert.issuer.rfc4514_string()}\n"
        certs_string += f"{cert.public_bytes(Encoding.PEM).decode()}\n"

    private_key_bytes: bytes = private_key.private_bytes(
        Encoding.PEM, PrivateFormat.PKCS8, serialization.NoEncryption()
    )

    return {
        "certData": base64.b64encode(certs_string.encode()).decode(),
        "privateKeyData": base64.b64encode(private_key_bytes).decode(),
        "createdAt": created_at,
        "certServiceURL": secret_data["certServiceURL"],
        "tokenURL": secret_data["tokenURL"],
        "clientID": secret_data["clientID"],
        "password": "",  # keep it empty to maintain structure, but indicate it's not needed
    }


def extract_message_data(pubsub_message: Any) -> Any:
    """Extracts secret rotation message from the Pub/Sub message."""
    if pubsub_message.get("attributes", {}).get("eventType") != "SECRET_ROTATE":
        raise ValueError("Unsupported event type")

    data = base64.b64decode(pubsub_message["data"])
    return json.loads(data)


def decrypt_private_key(private_key_data: bytes, password: bytes) -> bytes:
    """Decrypts an encrypted private key."""
    private_key: rsa.RSAPrivateKey = serialization.load_pem_private_key(
        private_key_data, password
    )

    return private_key.private_bytes(
        Encoding.PEM, PrivateFormat.PKCS8, serialization.NoEncryption()
    )


# TODO(kacpermalachowski): Move it to common package
def get_pubsub_message() -> Dict[str, Any]:
    """Parses the Pub/Sub message from the request."""
    envelope = request.get_json()
    if not envelope:
        raise ValueError("No Pub/Sub message received")

    if not isinstance(envelope, dict) or "message" not in envelope:
        raise ValueError("Invalid Pub/Sub message format")

    return envelope["message"]


# TODO(kacpermalachowski): Move it to common package
def prepare_error_response(err: str, logger: Logger) -> Response:
    """Prepares an error response with logging."""
    _, exc_value, _ = sys.exc_info()
    stacktrace = repr(traceback.format_exception(exc_value))
    logger.log_error(f"Error: {err}\nStack:\n {stacktrace}")
    resp = make_response()
    resp.content_type = "application/json"
    resp.status_code = 500
    return resp
