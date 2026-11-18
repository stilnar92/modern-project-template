# 0019 — A stolen copy of the database must not give working logins

From PR #106 · 12.11.2026 · Topics: security, sessions, hashing

## Problem

The server keeps a row for each login session ([lesson 0018](0018-sessions-or-tokens.md)). The token in the cookie works like a password: who has it is logged in. Copies of a database leave the server in many ways, for example a backup in the wrong place or a dump for a test. A copy must not be a list of working logins.

## Predict first

The `sessions` table keeps the token as plain text. Somebody gets a backup from last night. What can they do with one row? Now the table keeps only the SHA-256 hash of the token. What can they do?

## Theory

A **cryptographic hash** is a one-way function. From a value you can calculate its hash. From the hash you cannot calculate the value. The only attack is to guess values and hash each guess.

So the server can keep the hash, not the secret. When a token arrives, the server hashes it and searches for the hash. A person who reads the table sees hashes. A hash in a cookie does not work, because the server hashes it again and finds nothing.

Password storage uses the same idea, with one difference. People select passwords, so an attacker can guess them. A password therefore needs a slow hash with a salt (Argon2, bcrypt), which makes each guess expensive. A session token is a long random value from the server. Nobody can guess 128 random bits, so the speed of the hash does not help the attacker. A fast hash such as SHA-256 is sufficient, and the lookup on each request stays cheap.

The random value must come from a cryptographic generator. In Node.js terms: `crypto.randomBytes`, not `Math.random`.

The hash protects only against a person who can read the data. A person who can write to the table can insert the hash of a token of their own.

Sources:

- [OWASP Session Management Cheat Sheet](https://cheatsheetseries.owasp.org/cheatsheets/Session_Management_Cheat_Sheet.html): the length and the randomness of a session ID.
- [OWASP Password Storage Cheat Sheet](https://cheatsheetseries.owasp.org/cheatsheets/Password_Storage_Cheat_Sheet.html): why passwords need a slow hash.
- [The Copenhagen Book: Sessions](https://thecopenhagenbook.com/sessions): session tokens and why to store their hash.

## Options

| Option | How it works | Cost |
| --- | --- | --- |
| Keep the plain token | The lookup compares the cookie with the column | Each copy of the table is a list of working logins. The scs library does this |
| Keep a slow hash (bcrypt, Argon2) | As for passwords | Tens of milliseconds for each request, and no gain for a random token |
| Keep a fast hash (SHA-256) | Hash the cookie, search by the hash | A leak of the table gives no working token |
| Keep a hash with a server key (HMAC) | The hash also needs a secret that is not in the database | A key to keep and to change. It helps mainly when tokens are short |

## Edge cases

| What if | What happens |
| --- | --- |
| A backup of the database leaks | The copy has only hashes. No row gives a working cookie |
| An attacker sends a stolen hash as the cookie | The server hashes it again. The result matches no row |
| An attacker calls the lookup with guessed hashes | The token has 128 random bits, and the lookup needs the exact hash. It returns only an account ID, never a token |
| The server must show the token again later | It cannot. The token exists only in the cookie, from the moment the session starts |
| A wrong value goes into the hash column | The column accepts only 32 bytes, the size of a SHA-256 hash |
| An attacker steals the cookie from the browser, not the database | The hash does not help. `HttpOnly` and `Secure` protect the cookie, and logout ends the session |

## Decision

The cookie holds a random token of 128 bits. The table keeps only its SHA-256 hash, as the primary key, so the lookup is one index search. This was one reason not to use the scs library: it stores the plain token.

A fast hash would be wrong if the token were short or if a person selected it, for example a 6-digit login code. An attacker can try all 6-digit values against a hash in less than a second. Such a value needs a short lifetime and a limit on attempts, not only a hash.

Record: [ADR 0012](../../dev/adr/0012-infrastructure-libraries.md). The lookup function is in [ADR 0013](../../dev/adr/0013-session-lookup-function.md).

## In our code

1. `core/internal/accounts/accounts.go`: `StartSession` makes the token, and `tokenHash` is the only way to the database.
2. `core/migrations/00002_accounts_sessions.sql`: the `token_hash` column with its length check.
3. `core/internal/accounts/README.md`, section "Security": the rule in three lines.
4. [ADR 0013](../../dev/adr/0013-session-lookup-function.md), the stress round: guessed hashes.

## Check yourself

1. Somebody copies the `sessions` table. Can they log in as a user? Why?
2. Django hashes passwords with a slow function. Why is a fast SHA-256 correct for a session token?
3. Which attack does the hash not stop?

<details><summary>Answers</summary>

1. No. The table has only SHA-256 hashes. The cookie needs the token, and the hash does not give it back. A hash sent as a cookie gets hashed again and matches nothing.
2. A slow hash makes each guess expensive, which matters when people select weak passwords. The token is 128 random bits from the server. Guessing is impossible at any hash speed, so the slow function would only make each request slower.
3. The theft of the token itself, for example from the browser or from a network without TLS. It also does not stop a person who can write to the table: they can insert the hash of their own token.

</details>
