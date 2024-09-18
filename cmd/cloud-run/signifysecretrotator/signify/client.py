"""Custom client for signify API"""

import dataclasses
import json
import tempfile
from typing import Any
import requests
from cryptography.hazmat.primitives.asymmetric import rsa
from cryptography import x509
from cryptography.hazmat.primitives import serialization, hashes
from cryptography.hazmat.primitives.serialization import pkcs7


@dataclasses.dataclass
class CertificateSigningRequestConfig:
    """CSR configuration"""

    def __init__(
        self, validity_time: int, validity_unit: str, policy_name: str
    ) -> None:
        self.validity: dict[str, Any] = {"value": validity_time, "type": validity_unit}
        self.policy: str = policy_name


class SignifyClient:
    """Wraps signify API"""

    def __init__(
        self,
        token_url: str,
        certificate_service_url: str,
        client_id: str,
        csr_config: CertificateSigningRequestConfig = CertificateSigningRequestConfig(
            validity_time=7,
            validity_unit="DAYS",
            policy_name="sap-cloud-platform-clients",
        ),
    ) -> None:
        self.token_url: str = token_url
        self.certificate_service_url: str = certificate_service_url
        self.client_id: str = client_id
        self.csr_config: CertificateSigningRequestConfig = csr_config

    def fetch_access_token(
        self,
        certificate: bytes,
        private_key: bytes,
        timeout=30,
    ) -> str:
        """fetches access token from given token_url using certificate and private key"""
        # Use temporary file for old cert and key because requests library needs file paths,
        # the code is running in known environment controlled by us
        with (
            tempfile.NamedTemporaryFile() as old_cert_file,
            tempfile.NamedTemporaryFile() as old_key_file,
        ):

            old_cert_file.write(certificate)
            old_cert_file.flush()

            old_key_file.write(private_key)
            old_key_file.flush()

            access_token_response: requests.Response = requests.post(
                self.token_url,
                cert=(old_cert_file.name, old_key_file.name),
                data={
                    "grant_type": "client_credentials",
                    "client_id": self.client_id,
                },
                timeout=timeout,
            )

            if access_token_response.status_code != 200:
                raise requests.HTTPError(
                    f"Got non-success status code {access_token_response.status_code}",
                    response=access_token_response,
                )

            decoded_response = access_token_response.json()

            if "access_token" not in decoded_response:
                raise ValueError(
                    f"Got unexpected response structure: {decoded_response}"
                )

            return decoded_response["access_token"]

    def fetch_new_certificate(
        self, cert_data: bytes, private_key: rsa.RSAPrivateKey, access_token: str
    ) -> list[x509.Certificate]:
        """Fetch new certificates from given certificate service"""

        csr: x509.CertificateSigningRequest = self._prepare_csr(cert_data, private_key)

        crt_create_payload: str = self._prepare_cert_request_payload(csr)

        cert_create_response: requests.Response = requests.post(
            self.certificate_service_url,
            headers={
                "Authorization": f"Bearer {access_token}",
                "Content-Type": "application/json",
                "Accept": "application/json",
            },
            data=crt_create_payload,
            timeout=10,
        )

        if cert_create_response.status_code != 200:
            raise requests.HTTPError(
                f"Got non-success status code {cert_create_response.status_code}"
            )

        decoded_response = cert_create_response.json()

        if (
            "certificateChain" not in decoded_response
            or "value" not in decoded_response["certificateChain"]
        ):
            raise ValueError(
                f"Cannot issue new certifacte, invalid response format: {decoded_response}"
            )

        pkcs7_certs = decoded_response["certificateChain"]["value"].encode()

        return pkcs7.load_pem_pkcs7_certificates(pkcs7_certs)

    def _prepare_cert_request_payload(self, csr: x509.CertificateSigningRequest) -> str:
        return json.dumps(
            {
                "csr": {
                    "value": csr.public_bytes(serialization.Encoding.PEM).decode(
                        "utf-8"
                    )
                },
                "validity": self.csr_config.validity,
                "policy": self.csr_config.policy,
            }
        )

    def _prepare_csr(
        self, cert_data: bytes, private_key: rsa.RSAPrivateKey
    ) -> x509.CertificateSigningRequest:
        old_cert: x509.Certificate = x509.load_pem_x509_certificate(cert_data)

        csr: x509.CertificateSigningRequest = (
            x509.CertificateSigningRequestBuilder()
            .subject_name(old_cert.subject)
            .sign(private_key, hashes.SHA256())
        )

        return csr
