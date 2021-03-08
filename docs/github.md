# Connecting to GitHub

1. [Obtaining an Access Token](#Obtaining-an-Access-Token)
1. [Configuring your Environment](#Configuring-your-Environment)
1. [Running in Docker](#Running-in-Docker)
1. [Deploying to Kubernetes with Helm](#Deploying-to-Kubernetes-with-Helm)

## Obtaining an Access Token

Please see the GitHub guide for creating a [personal access token][].

[personal access token]: https://docs.github.com/en/free-pro-team@latest/github/authenticating-to-github/creating-a-personal-access-token

## Configuring your Environment

```bash
export GITHUB_USERNAME="username_for_token"
export GITHUB_ACCESS_TOKEN="access_token"
export GITHUB_ORGANIZATIONS="your_org[,another_org]"

# found on your account settings page: https://app.effx.com/account_settings
export EFFX_API_KEY="effx_api_key"
```

## Running in Docker

When running in docker, you'll need to pass along the various environment variables.

```bash
docker run --rm -it \
  -e GITHUB_USERNAME \
  -e GITHUB_ACCESS_TOKEN \
  -e GITHUB_ORGANIZATIONS \
  -e EFFX_API_KEY \
  effxhq/vcs-connect \
  github
```

Optionally you may also pass in a list of features you want to disable, such as
Language Detection.

```bash
-e DISABLE="LANGUAGE_DETECTION"
```


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
  name: github-vcs-connect
data:
  GITHUB_USERNAME: $(echo -n "${GITHUB_USERNAME}" | base64 | tr -d $'\n')
  GITHUB_ACCESS_TOKEN: $(echo -n "${GITHUB_ACCESS_TOKEN}" | base64 | tr -d $'\n')
  GITHUB_ORGANIZATIONS: $(echo -n "${GITHUB_ORGANIZATIONS}" | base64 | tr -d $'\n')
  EFFX_API_KEY: $(echo -n "${EFFX_API_KEY}" | base64 | tr -d $'\n')
EOF
```

Once the namespace and credentials have been setup, we can deploy vcs-connect.
Be sure to point your `externalConfig` at the proper secret.

```
helm upgrade -i github effxhq/vcs-connect \
  -n effx \
  --set provider=github \
  --set externalConfig.secretRef.name=github-vcs-connect
```

Once created, you can manually deploy a job to perform an initial indexing run.

```bash
kubectl create job -n effx --from cronjob/github-vcs-connect github-vcs-connect-$(date %s)
```
