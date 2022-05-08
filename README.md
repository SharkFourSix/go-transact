# go-transact
Turn bank transaction notifications into action.

## How it works
---

1. You receive money in your bank account.
2. Your institution sends a notification to your email address with transaction details. Importantly, the transaction has a description field which contains a unique string (let's call it a vendor reference id) 
3. Your email provider redirects the email to go-transact
4. go-transact extracts transaction information from the email
5. go-transact POSTS the transaction details to a callback url of your choosing

#### Use cases
---

- Can be used in environments where payment API services are non-existent or hard to acquire to automate service offerings and subscriptions upon receipt of payment.
- etc 👀

#### Data storage
---

A database (SQLite) file will be created in the current working directory to store the following:

- Received emails (unmatched emails are treated as spam)
- Callback status and data

[Jump to setup](#setup)

## Building
---

The provided [Makefile](Makefile) contains the necessary commans to build.

- To build a stripped down version: ```make build```
- To clean without cleaning build and test cache: ```make clean```
- To clean everything: ```make clean-all```

## Usage
---

### Requirements

1. Have an account at a bank that allows specifying a description when sending money.
2. Have the account receive email notifications upon receipt of credit transactions.
3. Email provider must be able to redirect emails.
4. Have a server to run _go-transact_ on.

### Setup

1. In your mail client, note your bank's email address and create a rule to redirect credit transaction emails to where **_go-transact_** will be running, i.e, `someuniqu_email@my-server.com`. I recommend using unguessable generated mailbox names such as `openssl rand -hex 24`
2. In your [config.yaml](config.yaml), configure the regex patterns that will be used to extract transaction information from the mail.
3. Run go-transact (prefereably as a service)
4. (**Optional but recommended**) [Setup firewall rules](#security-considerations) to only allow connections from mail service provider servers on port 25

To start daemon 

```shell
./go-transact --config-file myconfig.yaml
```

To show usage

```shell
./go-transact --help
```

## Security considerations
---

This program will only process emails from configured mailbox addresses to block uninvited visitors.

To prevent unwanted emails, you can setup firewall rules to only allow incoming connections from your mail provider.

The following example would only allow Outlook SMTP server to connect on port 25

```
iptables -A INPUT -s 40.92.18.77 -p tcp --dport 25 -j ACCEPT
iptables -A INPUT -p tcp --dport 25 -j DROP
```

## Changelog
---

### v1.1.1 | 2022-05-08

- Fix typo in struct field in [config.go](config/config.go) 

### v1.1.0 | 2022-05-08

- Automatic cleaning of amount values. 

### v1.0.6 | 2022-05-08

- Improvements

### v1.0.5 | 2022-05-08

- Improvements

### v1.0.4 | 2022-05-08

- Workflow improvements

### v1.0.3 | 2022-05-08

- Workflow improvements

### v1.0.2 | 2022-05-08

- Fix workflow write permissions

### v1.0.1 | 2022-05-08

- Small improvements

### v1.0.0 | 2022-05-08

- Initial release