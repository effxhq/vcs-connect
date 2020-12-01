# Connecting to GitLab

```bash
export GITLAB_USERNAME=""
export GITLAB_ACCESS_TOKEN=""
export GITLAB_GROUPS=""

docker run --rm -it \
  -e GITLAB_USERNAME \
  -e GITLAB_ACCESS_TOKEN \
  -e GITLAB_GROUPS \
  effxhq/vcs-connect \
  gitlab
```

## Deploying with Helm

TBD
