# Changelog

## 2.0.0

- add support for [AWS IAM authentication](https://www.vaultproject.io/docs/auth/aws.html) in Vault
  - vault approle auth is now optional (outside of docker) when other auth is enabled
- restructure vault configuration to have sub-levels for each type of authentication
  - vault auth backend path is now required


## 1.0.1

- add application version to the `/health-check` endpoint


## 1.0.0

- Initial open source release