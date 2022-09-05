# jpnic-admin-daemon
[![Publish Docker image(prod)](https://github.com/homenoc/jpnic-admin-daemon/actions/workflows/build-prod.yaml/badge.svg)](https://github.com/homenoc/jpnic-admin-daemon/actions/workflows/build-prod.yaml)
[![Publish Docker image(dev)](https://github.com/homenoc/jpnic-admin-daemon/actions/workflows/build-dev.yaml/badge.svg)](https://github.com/homenoc/jpnic-admin-daemon/actions/workflows/build-dev.yaml)

### Docker build
#### Produciton
```
docker build -t yoneyan/jpnic-daemon:production .
docker push yoneyan/jpnic-daemon:production
```
#### Develop
```
docker build -t yoneyan/jpnic-daemon:develop .
docker push yoneyan/jpnic-daemon:develop
```