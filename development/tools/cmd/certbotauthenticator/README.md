# certbotauthenticator

## Overview

Certbotauthenticator is a binary called by the certbot when it generates the certificate. The binary is used in during manual DNS challenge authentication. In the manual mode, the certbot passes the domain name and the authentication token as environment variables to the certbotauthenticator to create a TXT record in the domain. This way, the Let's Encrypt system can validate the domain ownership. After the validation completes, the certbotauthenticator is called again to clean the TXT records.

## Justification

We have created this tool because the certbot dns-google plugin does not support creating TXT records within a DNS zone existing in different project than the service account used to authenticate in GCP.

## Usage

### Environment variables

In manual mode, certbot passes this data as environment variables when calling external tools for authentication and clean-up.

- **CERTBOT_VALIDATION** - validation token, expected by Let's Enrypt as a record value.
- **CERTBOT_DOMAIN** - domain name against which Let's Encrypt will execute validation.
- **CERTBOT_AUTH_OUTPUT** - the output from the authentication step, passed only to call during the clean-up.

### Authentication
Use the **GOOGLE_APPLICATION_CREDENTIALS** environment variable to provide credentials for authentication in GCP.

### CLI parameters

The certbotauthenticator accepts the following command line parameters:

|Parameter | Description | Value|
|-----------|------------|-------|
| **-D** | When set to `true`, the record will be deleted. | `False` | 
