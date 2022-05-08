# go-transact

https://github.com/SharkFourSix/go-transact

## Usage

-- Requirements

1. Have an account at a bank that allows specifying a description when sending money.
2. Have the account receive email notifications upon receipt of credit transactions.
3. Email provider must be able to redirect emails.
4. Have a server to run _go-transact_ on.

-- Setup

1. In your mail client, note your bank's email address and create a rule to redirect 
    credit transaction emails to where **_go-transact_** will be running, 
    i.e, `someuniqu_email@my-server.com`. I recommend using an unguessable generated 
    mailbox names such as `openssl rand -hex 24`.

2. In your [config.yaml](config.yaml), configure the regex patterns that will be used 
    to extract transaction information from the mail.

3. Run go-transact (prefereably as a service).

4. (Optional but recommended) (See "Security considerations" section below) to only 
    allow connections from mail service provider servers on port 25

To start daemon 

./go-transact --config-file myconfig.yaml

## Security considerations

This program will only process emails from configured mailbox addresses to block uninvited visitors.

To prevent unwanted emails, you can setup firewall rules to only allow incoming connections from your mail provider.

The following example would only allow Outlook SMTP server to connect on port 25

iptables -A INPUT -s 40.92.18.77 -p tcp --dport 25 -j ACCEPT
iptables -A INPUT -p tcp --dport 25 -j DROP