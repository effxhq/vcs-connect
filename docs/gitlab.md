# Connecting to GitLab

1. [Obtaining an Access Token](#Obtaining-an-Access-Token)
1. [Configuring your Environment](#Configuring-your-Environment)
1. [Running in Docker](#Running-in-Docker)
1. [Deploying to Kubernetes with Helm](#Deploying-to-Kubernetes-with-Helm)

## Obtaining an Access Token

Please see the GitLab guide for creating a [personal access token][].

[personal access token]: https://docs.gitlab.com/ee/user/profile/personal_access_tokens.html

## Configuring your Environment

```bash
export GITLAB_USERNAME="username_for_token"
export GITLAB_ACCESS_TOKEN="access_token"
export GITLAB_GROUPS="your_group[,another_group]"

# found on your account settings page: https://app.effx.com/account_settings
export EFFX_API_KEY="effx_api_key"
```

## Running in Docker

When running in docker, you'll need to pass along the various environment variables.

```bash
docker run --rm -it \
  -e GITLAB_USERNAME \
  -e GITLAB_ACCESS_TOKEN \
  -e GITLAB_GROUPS \
  -e EFFX_API_KEY \
  effxhq/vcs-connect \
  gitlab
```

Optionally you may also pass in a list of features you want to disable, such as
Language Detection.

```bash
-e DISABLE="LANGUAGE_DETECTION"

## Deploying to Kubernetes with Helm

First, you'll need to add the effx helm repository.

```bash
helm repo add effxhq https://charts.effx.run
helm repo update
```

Before deploying the system, we'll first need to setup the namespace and credentials.

```bash
kubectl create ns effx

cat <<EOF | kubectl apply -f -
apiVersion: v1
kind: Secret
metadata:
  namespace: effx
  name: gitlab-vcs-connect
data:
  GITLAB_USERNAME: $(echo -n "${GITLAB_USERNAME}" | base64 | tr -d $'\n')
  GITLAB_ACCESS_TOKEN: $(echo -n "${GITLAB_ACCESS_TOKEN}" | base64 | tr -d $'\n')
  GITLAB_GROUPS: $(echo -n "${GITLAB_GROUPS}" | base64 | tr -d $'\n')
  EFFX_API_KEY: $(echo -n "${EFFX_API_KEY}" | base64 | tr -d $'\n')
EOF
```

Once the namespace and credentials have been setup, we can deploy vcs-connect.
Be sure to point your `externalConfig` at the proper secret.

```bash
helm upgrade -i gitlab effxhq/vcs-connect \
  -n effx \
  --set provider=gitlab \
  --set externalConfig.secretRef.name=gitlab-vcs-connect
```

Once created, you can manually deploy a job to perform an initial indexing run.

```bash
kubectl create job -n effx --from cronjob/gitlab-vcs-connect gitlab-vcs-connect-$(date %s)
```
