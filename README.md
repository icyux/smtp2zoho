smtp2zoho
===

`smtp2zoho` works as a proxy, receiving SMTP messages, forwarding mails to *Zoho Mail API*.

It is usually used to hide sender's IP address when using SMTP.

Purpose
---

In other words, it is a middleware between an app that sending mails via SMTP and *Zoho Mail API*, "updating" the app to support *Zoho Mail API* silently and non-invasively.

Undoubtedly, *Zoho* has been providing native SMTP access for all users. 
However, It is a problem that sending mail via SMTP could expose the IP address of sender. 
Receiver could check mail headers (it is easy as you want), which contains `Received` header, indicating sender IP address. 
When you are sending a mail by web app or HTTP API, IP of sender would not contained in headers. That is what it based on to avoid IP exposure.

Configure
---

Configure options are contained in "config.yml". Deploy your own Zoho access credential and specify fake SMTP server listen address.

### Create your Zoho client

First, create a "Self Client".
> [Setup OAuth with Zoho > Self Client](https://www.zoho.com/accounts/protocol/oauth-setup.html)

Get authorization code with this scope: "ZohoMail.accounts.READ,ZohoMail.messages.CREATE".
> [Kaizen #2 - OAuth2.0 and Self Client #API](https://help.zoho.com/portal/en/community/topic/kaizen-2-oauth2-0-and-self-client-api)

Now your have `Client ID`, `Client Secret` and `Authorization Code`.

### Get OAuth Token

Request this API to get `Access Token` and `Refresh Token`. (doc: https://www.zoho.com/mail/help/api/using-oauth-2.html)
```http
POST /oauth/v2/token
Host: accounts.zoho.com
Content-Type: application/x-www-form-urlencoded
```
form data:
```yaml
client_id: <ClientID>
client_secret: <ClientSecret>
grant_type: authorization_code
code: <AuthorizationCode>
```
It will response with `Access Token` and `Refresh Token`.

### Get Account ID

Request this API to get account ID. (doc: https://www.zoho.com/mail/help/api/get-user-account-details.html)
```http
GET /api/accounts
Host: mail.zoho.com
Authorization: Bearer <AccessToken>
```
It will response with user info, including `Account ID` (field name is "accountId").

### Fill into config

Fill the required credentials into "config.yml". The field "ZohoMailAddr" is the sender's mail address. Understand that passing sender address in SMTP context with "MAIL FROM" command is just used to make a standard SMTP process. It won't be contained in HTTP request.

### deploy SMTP listen address

Update the field "smtpListen".

Run server
---

Run it as other Go programs. No command arguments are required.

Worth Noticing
---

- **DO NOT** expose the SMTP server in public network. It is weak, non-standard SMTP session could make it panicked. And currently,  there's no authentication option provided. 
- If you want to specify sender display name, use `From` header in raw mail data. Here are some examples:
```text
# normal
From: Specified Name <someone@example.com>
# UTF-8 encoded
From: =?UTF-8?B?5Y+R6YCB6ICF?= <someone@example.com>
# if you don't want to specify it, just leave it blank
From: <someone@example.com>
```
- `Subject` header only allow formats below:
```text
Subject: TestSubject
Subject: =?UTF-8?B?VGVzdEhlYWRlcg==?=
```
- Make issues or pull requests if you want to improve the project.
