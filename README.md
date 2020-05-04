# github-rulla-nycklar


### Problem statement
In a World of cloud we need secrets in a form of service account sometimes. The also 
need to be rotated to improve security to lower the risk if a secret gets feet. 

In this case we are working with the integration inbetween google cloud platform 
and github.com. We like service account and we like to use them in github actions
as secrets. To now rotate the secrets and keep things secure, we don't want to 
manually update the secrets.

### Solution
To support the problem we have. We have chosen to rotate the secrets using this 
fancy program. This program them uploads the new version on the secret to a 
known secret name that is the same in every repo this program is managing.

limitations. 
* only working with one google project
* one repo means one service account in google cloud

### Usage

```bash

./_bin/github-rulla-nycklar \
    --github-key-file="<name of key>.private-key.pem" \
    --github-app-id=<id> \
    --github-install-id=<id> \
    --owner=<org> \
    --repo-to-email="test-foo=github-test-foo@<project id>.iam.gserviceaccount.com" \
    --repo-to-email="test-bar=github-test-bar@<project id>.iam.gserviceaccount.com" \
    --secret-name="SuperHemligSecret"
```

### Dev setup
```
$ go mod vendor
$ go mod download
```
