# github-rulla-nycklar

container image: `quay.io/shelman/github-rulla-nycklar`


### Problem statement
In a World of cloud we need secrets in a form of service account sometimes. They also 
need to be rotated to improve security to lower the risk if a secret gets feet. 

In this case we are working with the integration inbetween google cloud platform 
and github.com. We like service accounts and we like to use them in github actions
as secrets. To now rotate the secrets and keep things secure, we don't want to 
manually update the secrets.

### Solution
To support the problem we have. We have chosen to rotate the secrets using this 
fancy program. This program them uploads the new version on the secret to a 
known secret name that is the same in every repo this program is managing.
This secret is a google service account, using this service account you
have the option to get secrets directly from google secret manger [link](https://github.com/GoogleCloudPlatform/github-actions/tree/master/get-secretmanager-secrets)

limitations. 
* one repo means one service account in google cloud.
* the name of the service account secret will be the same in all repos
* only one key is handled and that is the service account.

### Usage
example
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

### Installing
The way the example runs the tool is to run it in and github action. To do this you need 
to do the following. 

###### Install it as a github app in the github org.
- Permission that needs to be assigned
- Get Github AppId, InstallID
- Generate a github key file (the pem file)

The app requires these permissions:

| Permission | Access |
| ---------- | ------ |
| Actions | Read-only |
| Contents | Read-only |
| Metadata | Read-only |
| Secrets | Read & write |


###### Get github install id
To find the install id on github go to `Org > Settings > Installed Github Apps > AppName > Configure` 
in the URL you can see the install ID `https://github.com/organizations/<ORG>/settings/installations/<install id>`


###### Create google service account key.
The service account can act in multiple google project. To allow this the 
service account need to have `Service Account Key Admin` to be allowed to create/delete/list 
service account keys.

- Create service account 
- Get Service Account key 
- Assign `Service Account Key Admin` in all projects that need to have keys rotated


###### Create a github repo that can host the github action
doing the actual setup to make it run. At this point i expect that we have the following
- google service account
- github AppId
- github InstallID
- github app private key

You can find a example github action here [here](example/schedule-action.yaml)
using this fill in the information and place it in `.github/workflows/`


###### Create Github secrets
In project were we run the github action we need secrets. Some secrets needs to be
base64 encoded see list. 

NOTE. to base64 encode a file and copy to osx clipboard `cat credentials.json | base64 | pbcopy`

- Secret name `GCP_PROJECT_ID` string of project id were the service account lives
- Secret name `ORG` string the github org that it's running in
- Secret name `INSTALL_ID` string the github app install id 
- Secret name `APP_ID` string the github app id
- Secret name `GCP_SA_KEY` base64 encoded content of the service account json
- Secret name `PRIVATE_KEY_PEM` base64 encoded content of the github private key pem



### Dev setup
```
$ go mod vendor
$ go mod download
```
