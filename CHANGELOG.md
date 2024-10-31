# Changelog

## 2.1.0

- update go to 1.23
- update all dependencies

## 2.0.0

- add support for [Vault AWS IAM authentication](https://www.vaultproject.io/docs/auth/aws.html) in Vault
- add support for [Vault Kubernetes authentication](https://www.vaultproject.io/docs/auth/kubernetes) in Vault
- add support for [Google Cloud IAM authentication](https://www.vaultproject.io/docs/auth/gcp) in Vault
- vault AppRole auth is now optional as other auth methods can be specified instead
- the docker container no longer verifies that any auth configuration is present besides the mount path
- restructure vault configuration to have sub-level for authentication
- restructure vault configuration to have sub-levels within authentication for each type of authentication
- the mount point of the auth backend must now be specified with as `VAULT_AUTH_MOUNT_PATH`


## 1.0.1

- add application version to the `/health-check` endpoint


## 1.0.0

- Initial open source release