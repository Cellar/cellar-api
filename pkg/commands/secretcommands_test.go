package commands_test

import (
	"cellar/pkg/commands"
	"cellar/pkg/mocks"
	"cellar/pkg/models"
	"cellar/testing/testhelpers"
	"github.com/golang/mock/gomock"
	"testing"
	"time"
)

func TestWhenCreatingASecretFromContent(t *testing.T) {
	ctrl := gomock.NewController(t)

	encryptedData := testhelpers.RandomId(t)

	encryption := mocks.NewMockEncryption(ctrl)
	encryption.EXPECT().
		Encrypt(gomock.Any()).
		Return(encryptedData, nil)

	dataStore := mocks.NewMockDataStore(ctrl)
	dataStore.EXPECT().
		WriteSecret(gomock.Any()).
		Return(nil)

	expectedDuration := time.Minute * 11
	expectedExpiration := time.Now().Add(expectedDuration).UTC()
	expectedSecret := models.Secret{
		Content:         []byte("Super Secret Test Content"),
		ContentType:     models.ContentTypeText,
		AccessLimit:     100,
		ExpirationEpoch: testhelpers.EpochFromNow(expectedDuration),
	}

	response, _, err := commands.CreateSecret(dataStore, encryption, expectedSecret)
	testhelpers.Ok(t, err)
	t.Run("should return ID", testhelpers.AssertF(len(response.ID) == 64, "expected ID length of 64 got: %d", len(response.ID)))
	t.Run("should return access count of zero", testhelpers.EqualsF(0, response.AccessCount))
	t.Run("should return access limit", testhelpers.EqualsF(expectedSecret.AccessLimit, response.AccessLimit))
	t.Run("should return content type", testhelpers.EqualsF(models.ContentType(expectedSecret.ContentType), response.ContentType))
	t.Run("should return expiration", testhelpers.EqualsF(expectedExpiration.Format("2006-01-02 15:04:05 UTC"), response.Expiration.Format()))
	t.Run("should encrypt content", func(t *testing.T) {
		encryption.EXPECT().
			Encrypt(expectedSecret.Content).
			Times(1)
	})
	t.Run("should write to database", func(t *testing.T) {
		dataStore.EXPECT().
			WriteSecret(gomock.Any()).
			Times(1)
	})
}

func TestWhenCreatingASecretFromFile(t *testing.T) {
	ctrl := gomock.NewController(t)

	encryptedData := testhelpers.RandomId(t)

	encryption := mocks.NewMockEncryption(ctrl)
	encryption.EXPECT().
		Encrypt(gomock.Any()).
		Return(encryptedData, nil)

	dataStore := mocks.NewMockDataStore(ctrl)
	dataStore.EXPECT().
		WriteSecret(gomock.Any()).
		Return(nil)

	expectedDuration := time.Minute * 11
	expectedExpiration := time.Now().Add(expectedDuration).UTC()
	expectedSecret := models.Secret{
		Content:         []byte("Super Secret Test Content"),
		ContentType:     models.ContentTypeFile,
		AccessLimit:     100,
		ExpirationEpoch: testhelpers.EpochFromNow(expectedDuration),
	}

	response, _, err := commands.CreateSecret(dataStore, encryption, expectedSecret)
	testhelpers.Ok(t, err)
	t.Run("should return ID", testhelpers.AssertF(len(response.ID) == 64, "expected ID length of 64 got: %d", len(response.ID)))
	t.Run("should return access count of zero", testhelpers.EqualsF(0, response.AccessCount))
	t.Run("should return access limit", testhelpers.EqualsF(expectedSecret.AccessLimit, response.AccessLimit))
	t.Run("should return content type", testhelpers.EqualsF(models.ContentType(expectedSecret.ContentType), response.ContentType))
	t.Run("should return expiration", testhelpers.EqualsF(expectedExpiration.Format("2006-01-02 15:04:05 UTC"), response.Expiration.Format()))
	t.Run("should encrypt content", func(t *testing.T) {
		encryption.EXPECT().
			Encrypt(expectedSecret.Content).
			Times(1)
	})
	t.Run("should write to database", func(t *testing.T) {
		dataStore.EXPECT().
			WriteSecret(gomock.Any()).
			Times(1)
	})
}

func TestWhenCreatingASecretWithTooShortExpiration(t *testing.T) {
	ctrl := gomock.NewController(t)

	encryptedData := testhelpers.RandomId(t)

	encryption := mocks.NewMockEncryption(ctrl)
	encryption.EXPECT().
		Encrypt(gomock.Any()).
		Return(encryptedData, nil)

	dataStore := mocks.NewMockDataStore(ctrl)
	dataStore.EXPECT().
		WriteSecret(gomock.Any()).
		Return(nil)

	expectedDuration := time.Minute * 9
	expirationEpoch := testhelpers.EpochFromNow(expectedDuration)

	content := "Super Secret Test Content"
	accessLimit := 100
	secretRequest := models.Secret{
		Content:         []byte(content),
		AccessLimit:     accessLimit,
		ExpirationEpoch: expirationEpoch,
	}

	_, isValidationError, err := commands.CreateSecret(dataStore, encryption, secretRequest)
	t.Run("should return validation error", testhelpers.EqualsF(true, isValidationError))
	t.Run("should return an error", testhelpers.AssertF(err != nil, "error should not be nil"))
	t.Run("should not to database", func(t *testing.T) {
		dataStore.EXPECT().
			WriteSecret(gomock.Any()).
			Times(0)
	})
}

func TestWhenAccessingASecret(t *testing.T) {
	ctrl := gomock.NewController(t)

	encryptedData := testhelpers.RandomId(t)

	secret := models.Secret{
		ID:              testhelpers.RandomId(t),
		CipherText:      encryptedData,
		ContentType:     models.ContentTypeText,
		AccessLimit:     100,
		ExpirationEpoch: testhelpers.EpochFromNow(time.Minute),
	}

	encryption := mocks.NewMockEncryption(ctrl)
	encryption.EXPECT().
		Decrypt(encryptedData).
		Return(secret.CipherText, nil)

	dataStore := mocks.NewMockDataStore(ctrl)
	dataStore.EXPECT().
		ReadSecret(secret.ID).
		Return(&secret)
	dataStore.EXPECT().
		IncreaseAccessCount(secret.ID).
		Return(int64(1), nil)

	response, err := commands.AccessSecret(dataStore, encryption, secret.ID)
	testhelpers.Ok(t, err)

	t.Run("should return ID", testhelpers.EqualsF(secret.ID, response.ID))
	t.Run("should return correct content", testhelpers.EqualsF(secret.CipherText, response.Content))
	t.Run("should decrypt content", func(t *testing.T) {
		encryption.EXPECT().
			Decrypt(encryptedData).
			Times(1)
	})
	t.Run("should access from database", func(t *testing.T) {
		dataStore.EXPECT().
			ReadSecret(secret.ID).
			Times(1)
	})
	t.Run("should increase access", func(t *testing.T) {
		dataStore.EXPECT().
			IncreaseAccessCount(secret.ID).
			Times(1)
	})
}

func TestWhenAccessingASecretThatDoesNotExist(t *testing.T) {
	ctrl := gomock.NewController(t)

	encryption := mocks.NewMockEncryption(ctrl)

	dataStore := mocks.NewMockDataStore(ctrl)
	dataStore.EXPECT().
		ReadSecret(gomock.Any()).
		Return(nil)

	response, err := commands.AccessSecret(dataStore, encryption, testhelpers.RandomId(t))
	testhelpers.Ok(t, err)
	t.Run("should return nil", testhelpers.IsNilF(response))
	t.Run("should not attempt to decrypt content", func(t *testing.T) {
		encryption.EXPECT().
			Decrypt(gomock.Any()).
			Times(0)
	})
	t.Run("should attempt access from datastore", func(t *testing.T) {
		dataStore.EXPECT().
			ReadSecret(gomock.Any()).
			Times(1)
	})
	t.Run("should not attempt to update access", func(t *testing.T) {
		dataStore.EXPECT().
			IncreaseAccessCount(gomock.Any()).
			Times(0)
	})
}

func TestWhenGettingSecretMetadata(t *testing.T) {
	ctrl := gomock.NewController(t)

	encryptedData := testhelpers.RandomId(t)

	secret := models.Secret{
		ID:              testhelpers.RandomId(t),
		CipherText:      encryptedData,
		ContentType:     models.ContentTypeText,
		AccessCount:     1,
		AccessLimit:     100,
		ExpirationEpoch: testhelpers.EpochFromNow(time.Minute),
	}

	dataStore := mocks.NewMockDataStore(ctrl)
	dataStore.EXPECT().
		ReadSecret(secret.ID).
		Return(&secret)

	response := commands.GetSecretMetadata(dataStore, secret.ID)
	t.Run("should return ID", testhelpers.EqualsF(secret.ID, response.ID))
	t.Run("should return access count", testhelpers.EqualsF(1, response.AccessCount))
	t.Run("should return access limit", testhelpers.EqualsF(secret.AccessLimit, response.AccessLimit))
	t.Run("should return Duration", testhelpers.EqualsF(secret.Expiration().Format(), response.Expiration.Format()))
	t.Run("should read from database", func(t *testing.T) {
		dataStore.EXPECT().
			ReadSecret(secret.ID).
			Times(1)
	})
	t.Run("should not handle access", func(t *testing.T) {
		dataStore.EXPECT().
			IncreaseAccessCount(gomock.Any()).
			Times(0)
	})
}

func TestWhenGettingSecretMetadataForSecretThatDoesNotExist(t *testing.T) {
	ctrl := gomock.NewController(t)

	dataStore := mocks.NewMockDataStore(ctrl)
	dataStore.EXPECT().
		ReadSecret(gomock.Any()).
		Return(nil)

	response := commands.GetSecretMetadata(dataStore, testhelpers.RandomId(t))
	t.Run("should return nil", testhelpers.IsNilF(response))
	t.Run("should attempt to read from database", func(t *testing.T) {
		dataStore.EXPECT().
			ReadSecret(gomock.Any()).
			Times(1)
	})
	t.Run("should not attempt to update access", func(t *testing.T) {
		dataStore.EXPECT().
			IncreaseAccessCount(gomock.Any()).
			Times(0)
	})
}
func TestWhenDeletingASecret(t *testing.T) {
	ctrl := gomock.NewController(t)

	secret := &models.Secret{
		ID: testhelpers.RandomId(t),
	}

	dataStore := mocks.NewMockDataStore(ctrl)
	dataStore.EXPECT().
		DeleteSecret(secret.ID).
		Return(true, nil)

	response, err := commands.DeleteSecret(dataStore, secret.ID)
	t.Run("should not return error", testhelpers.OkF(err))
	t.Run("should return true", testhelpers.EqualsF(true, response))
	t.Run("should delete from database", func(t *testing.T) {
		dataStore.EXPECT().
			DeleteSecret(secret.ID).
			Times(1)
	})
}

func TestWhenDeletingASecretThatDoesNotExist(t *testing.T) {
	ctrl := gomock.NewController(t)

	dataStore := mocks.NewMockDataStore(ctrl)
	dataStore.EXPECT().
		DeleteSecret(gomock.Any()).
		Return(false, nil)

	id := testhelpers.RandomId(t)

	response, err := commands.DeleteSecret(dataStore, id)
	t.Run("should not return error", testhelpers.OkF(err))
	t.Run("should return false", testhelpers.EqualsF(false, response))
	t.Run("should delete from database", func(t *testing.T) {
		dataStore.EXPECT().
			DeleteSecret(id).
			Times(1)
	})
}
