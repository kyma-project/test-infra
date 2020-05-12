# certbotauthenticator

## Overview

Certbotauthenticator is a binary called by the certbot when it generates the certificate. The binary is used in during manual DNS challenge authentication. In the manual mode, the certbot passes the domain name and the authentication token as environment variables to the certbotauthenticator to create a TXT record in the domain. This way, the Let's Encrypt system can validate the domain ownership. After the validation completes, the certbotauthenticator is called again to clean the TXT records.

##Justification

We made this tool, becasue certbot dns-google plugin doesn't support creating TXT records within DNS Zone existing in different project, than service account used to authenticate in GCP.

##Usage

###Environment variables

Certbot in manual mode when calling external tools for authentication and cleaning steps, is passing required data as environment variables.

- CERTBOT_VALIDATION - validation token, expected by LetsEnrypt as a record value.
- CERTBOT_DOMAIN - domain name against which LetsEncrypt will execute validation.
- CERTBOT_AUTH_OUTPUT - output from authentication step, it's passed only to call in cleaning step.

###Authentication

GOOGLE_APPLICATION_CREDENTIALS environment variable is used to provide credentials for authentication in GCP.

###CLI parameters

Certbotauthenticator is accepting following command line parameters.

- -D - When set to true, record will be deleted, default _false_
