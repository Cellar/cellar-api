package vault

import (
	pkgerrors "cellar/pkg/errors"
	"context"
)

func (vault EncryptionClient) login(ctx context.Context) error {
	if err := pkgerrors.CheckContext(ctx); err != nil {
		return err
	}

	vault.logger.Debug("attempting to find and renew existing tokens")
	token, err := vault.client.Auth().Token().RenewSelf(60)
	if err == nil && token != nil {
		vault.logger.Debug("token renewal successful")
		vault.client.SetToken(token.Auth.ClientToken)
		return nil
	} else {
		vault.logger.Debug("unable to find or renew existing tokens")
	}

	vault.logger.Debug("attempting to login to vault")
	authBackend, err := vault.configuration.AuthConfiguration()
	if err != nil {
		return err
	}
	loginParams, err := authBackend.LoginParameters()
	if err != nil {
		vault.logger.WithError(err).
			Error("unable to login to vault")
		return err
	}
	secret, err := vault.client.Logical().Write(authBackend.LoginPath(), loginParams)
	if err != nil {
		vault.logger.WithError(err).
			Error("unable to login to vault")
		return err
	}

	vault.logger.Debug("login to vault successful")
	vault.client.SetToken(secret.Auth.ClientToken)
	return nil
}
