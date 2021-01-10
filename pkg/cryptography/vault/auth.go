package vault

func (vault EncryptionClient) login() error {
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
	authBackend, err := vault.configuration.AuthBackend()
	if err != nil {
		return err
	}
	secret, err := vault.client.Logical().Write(authBackend.LoginPath(), authBackend.LoginParameters())
	if err != nil {
		vault.logger.WithError(err).
			Error("unable to login to vault")
		return err
	}

	vault.logger.Debug("login to vault successful")
	vault.client.SetToken(secret.Auth.ClientToken)
	return nil
}
