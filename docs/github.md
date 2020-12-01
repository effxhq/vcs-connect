# Connecting to GitHub

```bash
export GITHUB_USERNAME=""
export GITHUB_ACCESS_TOKEN=""
export GITHUB_ORGANIZATIONS=""

docker run --rm -it \
  -e GITHUB_USERNAME \
  -e GITHUB_ACCESS_TOKEN \
  -e GITHUB_ORGANIZATIONS \
  effxhq/vcs-connect \
  github
```

## Deploying with Helm

TBD
