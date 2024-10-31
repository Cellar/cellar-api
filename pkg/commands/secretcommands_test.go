package commands_test

import (
	"cellar/pkg/commands"
	"cellar/pkg/mocks"
	"cellar/pkg/models"
	"cellar/testing/testhelpers"
	"go.uber.org/mock/gomock"
	"testing"
	"time"
)

func TestWhenCreatingASecret(t *testing.T) {
	test := func(expectedContentType string) func(t *testing.T) {
		return func(t *testing.T) {
			encryptedData := testhelpers.RandomId(t)

			expectedDuration := time.Minute * 11
			expectedExpiration := time.Now().Add(expectedDuration).UTC()
			expectedSecret := models.Secret{
				Content:         []byte("Super Secret Test Content"),
				ContentType:     expectedContentType,
				AccessLimit:     100,
				ExpirationEpoch: testhelpers.EpochFromNow(expectedDuration),
			}

			sut := func(encryptCallTimes, writeSecretCallTimes int) (response *models.SecretMetadata) {
				ctrl := gomock.NewController(t)

				encryption := mocks.NewMockEncryption(ctrl)
				encryptCall := encryption.EXPECT().
					Encrypt(expectedSecret.Content).
					Return(encryptedData, nil).
					AnyTimes()

				if encryptCallTimes >= 0 {
					encryptCall.Times(encryptCallTimes)
				}

				dataStore := mocks.NewMockDataStore(ctrl)
				writeSecretCall := dataStore.EXPECT().
					WriteSecret(gomock.Any()).
					Return(nil).
					AnyTimes()

				if writeSecretCallTimes >= 0 {
					writeSecretCall.Times(writeSecretCallTimes)
				}

				response, _, err := commands.CreateSecret(dataStore, encryption, expectedSecret)
				testhelpers.Ok(t, err)

				return
			}

			t.Run("should encrypt content", func(t *testing.T) { sut(1, -1) })
			t.Run("should write to database", func(t *testing.T) { sut(-1, 1) })

			t.Run("should return", func(t *testing.T) {
				response := sut(-1, -1)

				t.Run("ID", testhelpers.AssertF(len(response.ID) == 64, "expected ID length of 64 got: %d", len(response.ID)))
				t.Run("access count of zero", testhelpers.EqualsF(0, response.AccessCount))
				t.Run("access limit", testhelpers.EqualsF(expectedSecret.AccessLimit, response.AccessLimit))
				t.Run("content type", testhelpers.EqualsF(models.ContentType(expectedSecret.ContentType), response.ContentType))
				t.Run("expiration", testhelpers.EqualsF(expectedExpiration.Format("2006-01-02 15:04:05 UTC"), response.Expiration.Format()))
			})
		}
	}

	t.Run("from content", test(models.ContentTypeText))
	t.Run("from file", test(models.ContentTypeFile))
}

func TestWhenCreatingASecretWithTooShortExpiration(t *testing.T) {

	encryptedData := testhelpers.RandomId(t)

	expectedDuration := time.Minute * 9
	expirationEpoch := testhelpers.EpochFromNow(expectedDuration)

	content := "Super Secret Test Content"
	accessLimit := 100

	secretRequest := models.Secret{
		Content:         []byte(content),
		AccessLimit:     accessLimit,
		ExpirationEpoch: expirationEpoch,
	}

	sut := func(writeSecretCallTimes int) (isValidationError bool, err error) {
		ctrl := gomock.NewController(t)

		encryption := mocks.NewMockEncryption(ctrl)
		encryption.EXPECT().
			Encrypt(gomock.Any()).
			Return(encryptedData, nil).
			AnyTimes()

		dataStore := mocks.NewMockDataStore(ctrl)
		writeSecretCall := dataStore.EXPECT().
			WriteSecret(gomock.Any()).
			Return(nil).
			AnyTimes()

		if writeSecretCallTimes >= 0 {
			writeSecretCall.Times(writeSecretCallTimes)
		}

		_, isValidationError, err = commands.CreateSecret(dataStore, encryption, secretRequest)

		return
	}

	t.Run("should not call to database", func(t *testing.T) { sut(0) })
	t.Run("should return", func(t *testing.T) {
		isValidationError, err := sut(-1)
		t.Run("should return validation error", testhelpers.EqualsF(true, isValidationError))
		t.Run("should return an error", testhelpers.AssertF(err != nil, "error should not be nil"))
	})
}

func TestWhenAccessingASecret(t *testing.T) {
	test := func(expectedContentType string) func(t *testing.T) {
		return func(t *testing.T) {

			secret := models.Secret{
				ID:              testhelpers.RandomId(t),
				Content:         []byte(testhelpers.RandomId(t)),
				CipherText:      testhelpers.RandomId(t),
				ContentType:     models.ContentTypeText,
				AccessLimit:     100,
				ExpirationEpoch: testhelpers.EpochFromNow(time.Minute),
			}

			sut := func(readSecretCallTimes, decryptCallTimes, increaseAccessCountCallTimes int) (response *models.Secret) {
				ctrl := gomock.NewController(t)

				encryption := mocks.NewMockEncryption(ctrl)
				decryptCall := encryption.EXPECT().
					Decrypt(secret.CipherText).
					Return(secret.Content, nil).
					AnyTimes()
				if decryptCallTimes >= 0 {
					decryptCall.Times(decryptCallTimes)
				}

				dataStore := mocks.NewMockDataStore(ctrl)
				readSecretCall := dataStore.EXPECT().
					ReadSecret(secret.ID).
					Return(&secret).
					AnyTimes()
				if readSecretCallTimes >= 0 {
					readSecretCall.Times(readSecretCallTimes)
				}

				increaseAccessCountCall := dataStore.EXPECT().
					IncreaseAccessCount(secret.ID).
					Return(int64(1), nil).
					AnyTimes()
				if increaseAccessCountCallTimes >= 0 {
					increaseAccessCountCall.Times(increaseAccessCountCallTimes)
				}

				response, err := commands.AccessSecret(dataStore, encryption, secret.ID)
				testhelpers.Ok(t, err)

				return
			}

			t.Run("should return", func(t *testing.T) {
				response := sut(-1, -1, -1)
				t.Run("should return ID", testhelpers.EqualsF(secret.ID, response.ID))
				t.Run("should return correct content", testhelpers.EqualsF(secret.Content, response.Content))
			})
			t.Run("should decrypt content", func(t *testing.T) { sut(-1, 1, -1) })
			t.Run("should access from database", func(t *testing.T) { sut(1, -1, -1) })
			t.Run("should increase access", func(t *testing.T) { sut(-1, -1, 1) })
		}
	}

	t.Run("from text", test(models.ContentTypeText))
	t.Run("from file", test(models.ContentTypeFile))
}

func TestWhenAccessingASecretThatDoesNotExist(t *testing.T) {

	sut := func(decryptCallTimes, readSecretCallTimes, increaseAccessCountCallTimes int) (response *models.Secret, err error) {
		ctrl := gomock.NewController(t)

		encryption := mocks.NewMockEncryption(ctrl)
		decryptCall := encryption.EXPECT().
			Decrypt(gomock.Any()).
			AnyTimes()
		if decryptCallTimes >= 0 {
			decryptCall.Times(decryptCallTimes)
		}

		dataStore := mocks.NewMockDataStore(ctrl)
		readSecretCall := dataStore.EXPECT().
			ReadSecret(gomock.Any()).
			Return(nil).
			AnyTimes()
		if readSecretCallTimes >= 0 {
			readSecretCall.Times(readSecretCallTimes)
		}
		increaseAccessCountCall := dataStore.EXPECT().
			IncreaseAccessCount(gomock.Any()).
			AnyTimes()
		if increaseAccessCountCallTimes >= 0 {
			increaseAccessCountCall.Times(increaseAccessCountCallTimes)
		}
		return commands.AccessSecret(dataStore, encryption, testhelpers.RandomId(t))
	}

	t.Run("should return", func(t *testing.T) {
		response, err := sut(-1, -1, -1)

		t.Run("should not return error", testhelpers.OkF(err))
		t.Run("should return nil", testhelpers.IsNilF(response))
	})

	t.Run("should not attempt to decrypt content", func(t *testing.T) { sut(0, -1, -1) })
	t.Run("should attempt access from datastore", func(t *testing.T) { sut(-1, 1, -1) })
	t.Run("should not attempt to update access", func(t *testing.T) { sut(-1, -1, 0) })
}

func TestWhenGettingSecretMetadata(t *testing.T) {
	encryptedData := testhelpers.RandomId(t)
	secret := models.Secret{
		ID:              testhelpers.RandomId(t),
		CipherText:      encryptedData,
		ContentType:     models.ContentTypeText,
		AccessCount:     1,
		AccessLimit:     100,
		ExpirationEpoch: testhelpers.EpochFromNow(time.Minute),
	}

	sut := func(readSecretCallTimes, increaseAccessCountCallTimes int) (response *models.SecretMetadata) {
		ctrl := gomock.NewController(t)

		dataStore := mocks.NewMockDataStore(ctrl)
		readSecretCall := dataStore.EXPECT().
			ReadSecret(secret.ID).
			Return(&secret).
			AnyTimes()

		if readSecretCallTimes >= 0 {
			readSecretCall.Times(readSecretCallTimes)
		}

		increaseAccessCountCall := dataStore.EXPECT().
			IncreaseAccessCount(gomock.Any()).
			Return(int64(1), nil).
			AnyTimes()

		if increaseAccessCountCallTimes >= 0 {
			increaseAccessCountCall.Times(increaseAccessCountCallTimes)
		}

		response = commands.GetSecretMetadata(dataStore, secret.ID)

		return
	}

	t.Run("should read from database", func(t *testing.T) {
		sut(1, -1)
	})
	t.Run("should not handle access", func(t *testing.T) {
		sut(-1, 0)
	})

	t.Run("should return", func(t *testing.T) {
		response := sut(-1, -1)
		t.Run("ID", testhelpers.EqualsF(secret.ID, response.ID))
		t.Run("access count", testhelpers.EqualsF(1, response.AccessCount))
		t.Run("access limit", testhelpers.EqualsF(secret.AccessLimit, response.AccessLimit))
		t.Run("Duration", testhelpers.EqualsF(secret.Expiration().Format(), response.Expiration.Format()))
	})

}

func TestWhenGettingSecretMetadataForSecretThatDoesNotExist(t *testing.T) {

	sut := func(readSecretCallTimes, increaseAccessCountCallTimes int) *models.SecretMetadata {
		ctrl := gomock.NewController(t)

		dataStore := mocks.NewMockDataStore(ctrl)
		readSecretCall := dataStore.EXPECT().
			ReadSecret(gomock.Any()).
			Return(nil).
			AnyTimes()
		if readSecretCallTimes >= 0 {
			readSecretCall.Times(readSecretCallTimes)
		}

		increaseAccessCountCall := dataStore.EXPECT().
			IncreaseAccessCount(gomock.Any()).
			AnyTimes()
		if increaseAccessCountCallTimes >= 0 {
			increaseAccessCountCall.Times(increaseAccessCountCallTimes)
		}

		return commands.GetSecretMetadata(dataStore, testhelpers.RandomId(t))
	}

	t.Run("should return nil", func(t *testing.T) {
		response := sut(-1, -1)
		testhelpers.IsNil(t, response)
	})
	t.Run("should attempt to read from database", func(t *testing.T) { sut(1, -1) })
	t.Run("should not attempt to update access", func(t *testing.T) { sut(-1, 0) })
}

func TestWhenDeletingASecret(t *testing.T) {
	sut := func(deleteSecretCallTimes int) (response bool, err error) {

		ctrl := gomock.NewController(t)

		secret := &models.Secret{
			ID: testhelpers.RandomId(t),
		}

		dataStore := mocks.NewMockDataStore(ctrl)
		deleteSecretCall := dataStore.EXPECT().
			DeleteSecret(secret.ID).
			Return(true, nil).
			AnyTimes()
		if deleteSecretCallTimes >= 0 {
			deleteSecretCall.Times(deleteSecretCallTimes)
		}

		return commands.DeleteSecret(dataStore, secret.ID)
	}

	t.Run("should return", func(t *testing.T) {
		response, err := sut(-1)
		t.Run("nil error", testhelpers.OkF(err))
		t.Run("true", testhelpers.EqualsF(true, response))
	})
	t.Run("should delete from database", func(t *testing.T) { sut(1) })
}

func TestWhenDeletingASecretThatDoesNotExist(t *testing.T) {
	sut := func(deleteSecretCallTimes int) (response bool, err error) {

		ctrl := gomock.NewController(t)

		dataStore := mocks.NewMockDataStore(ctrl)
		deleteSecretCall := dataStore.EXPECT().
			DeleteSecret(gomock.Any()).
			Return(false, nil).
			AnyTimes()
		if deleteSecretCallTimes >= 0 {
			deleteSecretCall.Times(deleteSecretCallTimes)
		}

		id := testhelpers.RandomId(t)

		return commands.DeleteSecret(dataStore, id)
	}

	t.Run("should return", func(t *testing.T) {
		response, err := sut(-1)
		t.Run("nil error", testhelpers.OkF(err))
		t.Run("false", testhelpers.EqualsF(false, response))
	})
	t.Run("should delete from database", func(t *testing.T) { sut(1) })
}
