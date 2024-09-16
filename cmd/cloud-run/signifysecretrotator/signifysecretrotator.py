"""Simple PubSub handler to rotate signify certificates"""

import datetime
import os
import base64
import json
import sys
import tempfile
import traceback
from typing import Any, Dict, List
import requests
from flask import Flask, Response, request, make_response
from cryptography import x509
from cryptography.hazmat.primitives import serialization, hashes
from cryptography.hazmat.primitives.serialization import pkcs7, Encoding, PrivateFormat
from cryptography.hazmat.primitives.asymmetric import rsa
from secretmanager import client

app = Flask(__name__)
project_id: str = os.getenv("PROJECT_ID")
component_name: str = os.getenv("COMPONENT_NAME", "signify-certificate-rotator")
application_name: str = os.getenv("APPLICATION_NAME", "secret-rotator")


# TODO(kacpermalachowski): Move it to common package
class LogEntry(dict):
    """Simplifies logging by returning a JSON string."""

    def __str__(self):
        return json.dumps(self)


@app.route("/", methods=["POST"])
def rotate_signify_secret() -> Response:
    """HTTP webhook handler for rotating Signify secrets."""
    log_fields: Dict[str, Any] = prepare_log_fields()
    log_fields["labels"]["io.kyma.app"] = "signify-certificate-rotate"

    try:
        sm_client = client.SecretManagerClient()

        if project_id is None:
            raise ValueError("Unknown project id")

        pubsub_message = get_pubsub_message()

        secret_rotate_msg = extract_message_data(pubsub_message)

        if secret_rotate_msg["labels"]["type"] != "signify":
            return prepare_error_response("Unsupported resource type", log_fields)

        secret_data = sm_client.get_secret(secret_rotate_msg["name"])

        old_cert_data = base64.b64decode(secret_data["certData"])
        old_pk_data = base64.b64decode(secret_data["privateKeyData"])

        if "password" in secret_data and secret_data["password"] != "":
            old_pk_data = decrypt_private_key(
                old_pk_data, secret_data["password"].encode()
            )

        new_private_key = rsa.generate_private_key(public_exponent=65537, key_size=4096)

        access_token = fetch_access_token(
            old_cert_data, old_pk_data, secret_data["tokenURL"], secret_data["clientID"]
        )

        created_at = datetime.datetime.now().strftime("%d-%m-%Y %H:%M:%S")

        new_certs: List[x509.Certificate] = fetch_new_certificate(
            old_cert_data, new_private_key, access_token, secret_data["certServiceURL"]
        )

        new_secret_data = prepare_new_secret(
            new_certs, new_private_key, secret_data, created_at
        )

        set_secret(
            secret_id=secret_rotate_msg["name"], data=json.dumps(new_secret_data)
        )

        print(
            LogEntry(
                severity="INFO",
                message="Certificate rotated successfully",
                **log_fields,
            )
        )

        return "Certificate rotated successfully"
    except ValueError as exc:
        return prepare_error_response(exc, log_fields)


def fetch_new_certificate(
    cert_data: bytes,
    private_key: rsa.RSAPrivateKey,
    access_token: str,
    certificate_service_url: str,
):
    """Fetch new certificates from given certificate service"""
    old_cert = x509.load_pem_x509_certificate(cert_data)

    csr = (
        x509.CertificateSigningRequestBuilder()
        .subject_name(old_cert.subject)
        .sign(private_key, hashes.SHA256())
    )

    crt_create_payload = json.dumps(
        {
            "csr": {
                "value": csr.public_bytes(serialization.Encoding.PEM).decode("utf-8")
            },
            "validity": {"value": 7, "type": "DAYS"},
            "policy": "sap-cloud-platform-clients",
        }
    )

    cert_create_response = requests.post(
        certificate_service_url,
        headers={
            "Authorization": f"Bearer {access_token}",
            "Content-Type": "application/json",
            "Accept": "application/json",
        },
        data=crt_create_payload,
        timeout=10,
    ).json()

    pkcs7_certs = cert_create_response["certificateChain"]["value"].encode()

    return pkcs7.load_pem_pkcs7_certificates(pkcs7_certs)


def prepare_new_secret(
    certificates: List[x509.Certificate],
    private_key: rsa.RSAPrivateKey,
    secret_data: Dict[str, Any],
    created_at: str,
) -> Dict[str, Any]:
    """Prepares new secret data with updated certificates and private key."""

    # format certificates
    certs_string = ""

    for cert in certificates:
        certs_string += f"subject={cert.subject.rfc4514_string()}\n"
        certs_string += f"issuer={cert.issuer.rfc4514_string()}\n"
        certs_string += f"{cert.public_bytes(Encoding.PEM).decode()}\n"

    private_key_bytes = private_key.private_bytes(
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
    # pylint: disable=line-too-long
    private_key = serialization.load_pem_private_key(private_key_data, password)

    # pylint: disable=line-too-long
    return private_key.private_bytes(
        Encoding.PEM, PrivateFormat.PKCS8, serialization.NoEncryption()
    )


def fetch_access_token(
    certificate: bytes, private_key: bytes, token_url: str, client_id: str
) -> str:
    """fetches access token from given token_url using certificate and private key"""
    # Use temporary file for old cert and key because requests library needs file paths,
    # it's not a security concern because the code is running in known environment controlled by us
    # pylint: disable=line-too-long
    with tempfile.NamedTemporaryFile() as old_cert_file, tempfile.NamedTemporaryFile() as old_key_file:

        old_cert_file.write(certificate)
        old_cert_file.flush()

        old_key_file.write(private_key)
        old_key_file.flush()

        # pylint: disable=line-too-long
        access_token_response = requests.post(
            token_url,
            cert=(old_cert_file.name, old_key_file.name),
            data={
                "grant_type": "client_credentials",
                "client_id": client_id,
            },
            timeout=30,
        ).json()

        return access_token_response["access_token"]


# TODO(kacpermalachowski): Move it to common package
def prepare_log_fields() -> Dict[str, Any]:
    """prepare_log_fields prapares basic log fields"""
    log_fields: Dict[str, Any] = {}
    request_is_defined = "request" in globals() or "request" in locals()
    if request_is_defined and request:
        trace_header = request.headers.get("X-Cloud-Trace-Context")
        if trace_header and project_id:
            trace = trace_header.split("/")
            log_fields["logging.googleapis.com/trace"] = (
                f"projects/{project_id}/traces/{trace[0]}"
            )
    log_fields["Component"] = "signify-certificate-rotator"
    log_fields["labels"] = {"io.kyma.component": "signify-certificate-rotator"}
    return log_fields


# TODO(kacpermalachowski): Move it to common package
def get_pubsub_message():
    """Parses the Pub/Sub message from the request."""
    envelope = request.get_json()
    if not envelope:
        # pylint: disable=broad-exception-raised
        raise ValueError("No Pub/Sub message received")

    if not isinstance(envelope, dict) or "message" not in envelope:
        # pylint: disable=broad-exception-raised
        raise ValueError("Invalid Pub/Sub message format")

    return envelope["message"]


# TODO(kacpermalachowski): Move it to common package
def prepare_error_response(err: str, log_fields: Dict[str, Any]) -> Response:
    """Prepares an error response with logging."""
    _, exc_value, _ = sys.exc_info()
    stacktrace = repr(traceback.format_exception(exc_value))
    print(
        LogEntry(
            severity="ERROR",
            message=f"Error: {err}\nStack:\n {stacktrace}",
            **log_fields,
        )
    )
    resp = make_response()
    resp.content_type = "application/json"
    resp.status_code = 500
    return resp


def set_secret(secret_id: str, data: str):
    """Adds a new version of the secret in Secret Manager."""
    pass

def setup_app():
    print("test")
    pass