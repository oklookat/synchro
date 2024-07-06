# Server idea. Not work now.

## Syntax

Requests/reponses uses json commands.

Request:

```json
{
  "command": "", // command name
  "data": null // object with command data
}
```

Response ok:

```json
{
  "command": "", // command name
  "data": null // object with response command data
}
```

Response error:

```json
{
  "command": "", // command name
  "error": "" // object with error message
}
```

## Entities

### Account

```json
{
  "id": "",
  "remoteName": "",
  "alias": "",
  "created_at": null // UNIX time
}
```

### RemoteName

- deezer
- spotify
- vkmusic
- yandexmusic
- zvuk

### Config (example)

```json
{
  "deezer": {
    "host": "http://localhost",
    "port": 8081
  },
  "general": {
    "debug": true
  },
  "linker": {
    "recheckMissing": false
  },
  "spotify": {
    "host": "http://localhost",
    "port": 8080
  },
  "yandexMusic": {
    "deviceID": "01J1WFVPBJ4SQKFS56JE5JZ440"
  }
}
```

## Commands

### cancel

Cancel command.

Request data:

```json
{
  "command": "" // command name
}
```

### configget

Get Config.

Response data: Config.

### configset

Set Config.

Request data: Config.

### accountlist

Get accounts list.

Request data:

```json
{
  "limit": 10,
  "offset": 0
}
```

Response data:

```json
[
  // Account objects.
]
```

### accountdelete

Delete account.

Request data:

```json
{
  "id": "" // account id
}
```

### accountalias

Set account alias.

Request data:

```json
{
  "id": "", // account id
  "alias": "" // new alias
}
```

### accountadd

Add streaming account.

Request data (auth):

```json
{
  "alias": "", // can be empty / null
  "remoteName": "", // remoteName
  "data": null // object with auth data, depends on streaming. See streamings below
}
```

Request data (reauth). Same flow as auth, but with existing account:

```json
{
  "id": "", // account id
  "remoteName": "",
  "data": null
}
```

Auth response depends on streaming. See streamings below.

p.s: but anyway, if auth ok, final response will contain new Account.

#### Deezer, Spotify

1. Auth data:

```json
{
  "appId": "",
  "appSecret": ""
}
```

2. Response data:

```json
{
  "url": "" // deezer auth url
}
```

User goes to URL, and confirms auth.

#### VK Music

1. Auth data:

```json
{
  "phone": "", // VK phone number
  "password": "" // VK password
}
```

2. Response data:

```json
{
  "current": "", // 2FA code method
  "resend": "" // alternative 2FA code method
}
```

On the "current" method (e.g. email), a code from VK is sent to the user. The "resend" method can have an alternative method (e.g. phone). The "resend" can be empty.

Ask the user to enter the code. Also inform user that the code can be sent via "resend".

3. Auth data:

```json
{
  "code": "", // user code. If resend == true, can be null/empty
  "resend": false // Resend code?
}
```

If resend code == true, auth goes to step 2 again.

If resend == false, and code valid, you authorized.

I was not able to fully replicate the VK login, because of which the token update method is not implemented. So sooner or later the token will expire and you need to reauth.

#### Yandex.Music

1. Auth data:

```json
null
```

2. Response data:

```json
{
  "url": "",
  "code": ""
}
```

User must go to url, and enter code. If ok, you authorized.

#### Zvuk

1. Auth data:

```json
{
  "token": "" // token from https://zvuk.com/api/tiny/profile
}
```

You authorized.

Due to the specifics of logging in via Sber ID, only this method is available. Accordingly, no checks are performed and the correctness of token entry depends only on the user. Plus, it will need to be updated periodically (via reauth).
