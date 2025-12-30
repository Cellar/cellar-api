package commands_test

import (
	"cellar/pkg/commands"
	pkgerrors "cellar/pkg/errors"
	"cellar/pkg/mocks"
	"cellar/pkg/models"
	"cellar/testing/testhelpers"
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestCreateSecret(t *testing.T) {
	t.Run("when all parameters are valid", func(t *testing.T) {
		test := func(expectedContentType string) func(t *testing.T) {
			return func(t *testing.T) {
				encryptedData := testhelpers.RandomId(t)

				expectedDuration := time.Minute * 11
				expectedExpiration := time.Now().Add(expectedDuration).UTC()
				maxAccessCount := 100
				maxExpirationSeconds := 604800

				expectedSecret := models.Secret{
					Content:         []byte("Super Secret Test Content"),
					ContentType:     expectedContentType,
					AccessLimit:     50,
					ExpirationEpoch: testhelpers.EpochFromNow(expectedDuration),
				}

				sut := func(encryptCallTimes, writeSecretCallTimes int) (response *models.SecretMetadata) {
					ctrl := gomock.NewController(t)

					encryption := mocks.NewMockEncryption(ctrl)
					encryptCall := encryption.EXPECT().
						Encrypt(gomock.Any(), expectedSecret.Content).
						Return(encryptedData, nil).
						AnyTimes()

					if encryptCallTimes >= 0 {
						encryptCall.Times(encryptCallTimes)
					}

					dataStore := mocks.NewMockDataStore(ctrl)
					writeSecretCall := dataStore.EXPECT().
						WriteSecret(gomock.Any(), gomock.Any()).
						Return(nil).
						AnyTimes()

					if writeSecretCallTimes >= 0 {
						writeSecretCall.Times(writeSecretCallTimes)
					}

					appConfig := mocks.NewMockIAppConfiguration(ctrl)
					appConfig.EXPECT().MaxAccessCount().Return(maxAccessCount).AnyTimes()
					appConfig.EXPECT().MaxExpirationSeconds().Return(maxExpirationSeconds).AnyTimes()

					response, err := commands.CreateSecret(context.Background(), appConfig, dataStore, encryption, expectedSecret)
					require.NoError(t, err)

					return
				}

				t.Run("it should encrypt content", func(t *testing.T) { sut(1, -1) })
				t.Run("it should write to database", func(t *testing.T) { sut(-1, 1) })

				t.Run("it should return", func(t *testing.T) {
					response := sut(-1, -1)

					t.Run("ID of length 64", func(t *testing.T) {
						assert.Equal(t, 64, len(response.ID))
					})

					t.Run("access count of zero", func(t *testing.T) {
						assert.Equal(t, 0, response.AccessCount)
					})

					t.Run("expected access limit", func(t *testing.T) {
						assert.Equal(t, expectedSecret.AccessLimit, response.AccessLimit)
					})

					t.Run("expected content type", func(t *testing.T) {
						assert.Equal(t, models.ContentType(expectedSecret.ContentType), response.ContentType)
					})

					t.Run("expected expiration", func(t *testing.T) {
						assert.Equal(t, expectedExpiration.Format("2006-01-02 15:04:05 UTC"), response.Expiration.Format())
					})
				})

				t.Run("when context is cancelled", func(t *testing.T) {
					t.Run("it should return context error", func(t *testing.T) {
						ctx, cancel := context.WithCancel(context.Background())
						cancel()

						ctrl := gomock.NewController(t)
						encryption := mocks.NewMockEncryption(ctrl)
						dataStore := mocks.NewMockDataStore(ctrl)
						appConfig := mocks.NewMockIAppConfiguration(ctrl)

						_, err := commands.CreateSecret(ctx, appConfig, dataStore, encryption, expectedSecret)

						assert.True(t, pkgerrors.IsContextError(err), "expected context error")
					})
				})
			}
		}

		t.Run("and from content", test(models.ContentTypeText))
		t.Run("and from file", test(models.ContentTypeFile))
	})

	t.Run("when expiration is too short", func(t *testing.T) {
		encryptedData := testhelpers.RandomId(t)

		expectedDuration := time.Minute * 9
		expirationEpoch := testhelpers.EpochFromNow(expectedDuration)
		maxAccessCount := 100
		maxExpirationSeconds := 604800

		content := "Super Secret Test Content"
		accessLimit := 50

		secretRequest := models.Secret{
			Content:         []byte(content),
			AccessLimit:     accessLimit,
			ExpirationEpoch: expirationEpoch,
		}

		sut := func(writeSecretCallTimes int) error {
			ctrl := gomock.NewController(t)

			encryption := mocks.NewMockEncryption(ctrl)
			encryption.EXPECT().
				Encrypt(gomock.Any(), gomock.Any()).
				Return(encryptedData, nil).
				AnyTimes()

			dataStore := mocks.NewMockDataStore(ctrl)
			writeSecretCall := dataStore.EXPECT().
				WriteSecret(gomock.Any(), gomock.Any()).
				Return(nil).
				AnyTimes()

			if writeSecretCallTimes >= 0 {
				writeSecretCall.Times(writeSecretCallTimes)
			}

			appConfig := mocks.NewMockIAppConfiguration(ctrl)
			appConfig.EXPECT().MaxAccessCount().Return(maxAccessCount).AnyTimes()
			appConfig.EXPECT().MaxExpirationSeconds().Return(maxExpirationSeconds).AnyTimes()

			_, err := commands.CreateSecret(context.Background(), appConfig, dataStore, encryption, secretRequest)
			return err
		}

		t.Run("it should not call to database", func(t *testing.T) { _ = sut(0) })
		t.Run("it should return", func(t *testing.T) {
			err := sut(-1)

			t.Run("validation error", func(t *testing.T) {
				assert.True(t, pkgerrors.IsValidationError(err))
			})

			t.Run("an error", func(t *testing.T) {
				assert.Error(t, err)
			})
		})
	})

	t.Run("when access limit exceeds maximum configured value", func(t *testing.T) {
		encryptedData := testhelpers.RandomId(t)

		expectedDuration := time.Hour
		expirationEpoch := testhelpers.EpochFromNow(expectedDuration)
		maxAccessCount := 100
		maxExpirationSeconds := 604800
		exceedingAccessLimit := maxAccessCount + 1

		content := "Super Secret Test Content"

		secretRequest := models.Secret{
			Content:         []byte(content),
			AccessLimit:     exceedingAccessLimit,
			ExpirationEpoch: expirationEpoch,
		}

		sut := func(writeSecretCallTimes int) error {
			ctrl := gomock.NewController(t)

			encryption := mocks.NewMockEncryption(ctrl)
			encryption.EXPECT().
				Encrypt(gomock.Any(), gomock.Any()).
				Return(encryptedData, nil).
				AnyTimes()

			dataStore := mocks.NewMockDataStore(ctrl)
			writeSecretCall := dataStore.EXPECT().
				WriteSecret(gomock.Any(), gomock.Any()).
				Return(nil).
				AnyTimes()

			if writeSecretCallTimes >= 0 {
				writeSecretCall.Times(writeSecretCallTimes)
			}

			appConfig := mocks.NewMockIAppConfiguration(ctrl)
			appConfig.EXPECT().
				MaxAccessCount().
				Return(maxAccessCount).
				AnyTimes()
			appConfig.EXPECT().
				MaxExpirationSeconds().
				Return(maxExpirationSeconds).
				AnyTimes()

			_, err := commands.CreateSecret(context.Background(), appConfig, dataStore, encryption, secretRequest)
			return err
		}

		t.Run("it should not call to database", func(t *testing.T) { _ = sut(0) })
		t.Run("it should return", func(t *testing.T) {
			err := sut(-1)

			t.Run("validation error", func(t *testing.T) {
				assert.True(t, pkgerrors.IsValidationError(err))
			})

			t.Run("an error", func(t *testing.T) {
				assert.Error(t, err)
			})
		})
	})

	t.Run("when expiration exceeds maximum configured value", func(t *testing.T) {
		encryptedData := testhelpers.RandomId(t)

		maxAccessCount := 100
		maxExpirationSeconds := 604800 // 7 days
		exceedingExpiration := time.Now().Add(time.Second * time.Duration(maxExpirationSeconds+1)).UTC()
		expirationEpoch := exceedingExpiration.Unix()

		content := "Super Secret Test Content"
		accessLimit := 10

		secretRequest := models.Secret{
			Content:         []byte(content),
			AccessLimit:     accessLimit,
			ExpirationEpoch: expirationEpoch,
		}

		sut := func(writeSecretCallTimes int) error {
			ctrl := gomock.NewController(t)

			encryption := mocks.NewMockEncryption(ctrl)
			encryption.EXPECT().
				Encrypt(gomock.Any(), gomock.Any()).
				Return(encryptedData, nil).
				AnyTimes()

			dataStore := mocks.NewMockDataStore(ctrl)
			writeSecretCall := dataStore.EXPECT().
				WriteSecret(gomock.Any(), gomock.Any()).
				Return(nil).
				AnyTimes()

			if writeSecretCallTimes >= 0 {
				writeSecretCall.Times(writeSecretCallTimes)
			}

			appConfig := mocks.NewMockIAppConfiguration(ctrl)
			appConfig.EXPECT().
				MaxAccessCount().
				Return(maxAccessCount).
				AnyTimes()
			appConfig.EXPECT().
				MaxExpirationSeconds().
				Return(maxExpirationSeconds).
				AnyTimes()

			_, err := commands.CreateSecret(context.Background(), appConfig, dataStore, encryption, secretRequest)
			return err
		}

		t.Run("it should not call to database", func(t *testing.T) { _ = sut(0) })
		t.Run("it should return", func(t *testing.T) {
			err := sut(-1)

			t.Run("validation error", func(t *testing.T) {
				assert.True(t, pkgerrors.IsValidationError(err))
			})

			t.Run("an error", func(t *testing.T) {
				assert.Error(t, err)
			})
		})
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
					Decrypt(gomock.Any(), secret.CipherText).
					Return(secret.Content, nil).
					AnyTimes()
				if decryptCallTimes >= 0 {
					decryptCall.Times(decryptCallTimes)
				}

				dataStore := mocks.NewMockDataStore(ctrl)
				readSecretCall := dataStore.EXPECT().
					ReadSecret(gomock.Any(), secret.ID).
					Return(&secret).
					AnyTimes()
				if readSecretCallTimes >= 0 {
					readSecretCall.Times(readSecretCallTimes)
				}

				increaseAccessCountCall := dataStore.EXPECT().
					IncreaseAccessCount(gomock.Any(), secret.ID).
					Return(int64(1), nil).
					AnyTimes()
				if increaseAccessCountCallTimes >= 0 {
					increaseAccessCountCall.Times(increaseAccessCountCallTimes)
				}

				response, err := commands.AccessSecret(context.Background(), dataStore, encryption, secret.ID)
				require.NoError(t, err)

				return
			}

			t.Run("should return", func(t *testing.T) {
				response := sut(-1, -1, -1)

				t.Run("it should return ID", func(t *testing.T) {
					assert.Equal(t, secret.ID, response.ID)
				})

				t.Run("it should return correct content", func(t *testing.T) {
					assert.Equal(t, secret.Content, response.Content)
				})
			})
			t.Run("should decrypt content", func(t *testing.T) { sut(-1, 1, -1) })
			t.Run("should access from database", func(t *testing.T) { sut(1, -1, -1) })
			t.Run("should increase access", func(t *testing.T) { sut(-1, -1, 1) })

			t.Run("when context is cancelled", func(t *testing.T) {
				t.Run("it should return context error", func(t *testing.T) {
					ctx, cancel := context.WithCancel(context.Background())
					cancel()

					ctrl := gomock.NewController(t)
					encryption := mocks.NewMockEncryption(ctrl)
					dataStore := mocks.NewMockDataStore(ctrl)

					_, err := commands.AccessSecret(ctx, dataStore, encryption, secret.ID)

					assert.True(t, pkgerrors.IsContextError(err), "expected context error")
				})
			})
		}
	}

	t.Run("from text", test(models.ContentTypeText))
	t.Run("from file", test(models.ContentTypeFile))
}

func TestWhenAccessingASecretWithFilename(t *testing.T) {
	expectedFilename := "test-document.pdf"

	secret := models.Secret{
		ID:              testhelpers.RandomId(t),
		Content:         []byte(testhelpers.RandomId(t)),
		CipherText:      testhelpers.RandomId(t),
		ContentType:     models.ContentTypeFile,
		Filename:        expectedFilename,
		AccessLimit:     100,
		ExpirationEpoch: testhelpers.EpochFromNow(time.Minute),
	}

	ctrl := gomock.NewController(t)

	encryption := mocks.NewMockEncryption(ctrl)
	encryption.EXPECT().
		Decrypt(gomock.Any(), secret.CipherText).
		Return(secret.Content, nil)

	dataStore := mocks.NewMockDataStore(ctrl)
	dataStore.EXPECT().
		ReadSecret(gomock.Any(), secret.ID).
		Return(&secret)
	dataStore.EXPECT().
		IncreaseAccessCount(gomock.Any(), secret.ID).
		Return(int64(1), nil)

	response, err := commands.AccessSecret(context.Background(), dataStore, encryption, secret.ID)
	require.NoError(t, err)

	t.Run("it should return filename", func(t *testing.T) {
		assert.Equal(t, expectedFilename, response.Filename)
	})
}

func TestWhenAccessingASecretThatDoesNotExist(t *testing.T) {

	sut := func(decryptCallTimes, readSecretCallTimes, increaseAccessCountCallTimes int) (response *models.Secret, err error) {
		ctrl := gomock.NewController(t)

		encryption := mocks.NewMockEncryption(ctrl)
		decryptCall := encryption.EXPECT().
			Decrypt(gomock.Any(), gomock.Any()).
			AnyTimes()
		if decryptCallTimes >= 0 {
			decryptCall.Times(decryptCallTimes)
		}

		dataStore := mocks.NewMockDataStore(ctrl)
		readSecretCall := dataStore.EXPECT().
			ReadSecret(gomock.Any(), gomock.Any()).
			Return(nil).
			AnyTimes()
		if readSecretCallTimes >= 0 {
			readSecretCall.Times(readSecretCallTimes)
		}
		increaseAccessCountCall := dataStore.EXPECT().
			IncreaseAccessCount(gomock.Any(), gomock.Any()).
			AnyTimes()
		if increaseAccessCountCallTimes >= 0 {
			increaseAccessCountCall.Times(increaseAccessCountCallTimes)
		}
		return commands.AccessSecret(context.Background(), dataStore, encryption, testhelpers.RandomId(t))
	}

	t.Run("should return", func(t *testing.T) {
		response, err := sut(-1, -1, -1)

		t.Run("it should not return error", func(t *testing.T) {
			assert.NoError(t, err)
		})

		t.Run("it should return nil", func(t *testing.T) {
			assert.Nil(t, response)
		})
	})

	t.Run("should not attempt to decrypt content", func(t *testing.T) { _, _ = sut(0, -1, -1) })
	t.Run("should attempt access from datastore", func(t *testing.T) { _, _ = sut(-1, 1, -1) })
	t.Run("should not attempt to update access", func(t *testing.T) { _, _ = sut(-1, -1, 0) })
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
			ReadSecret(gomock.Any(), secret.ID).
			Return(&secret).
			AnyTimes()

		if readSecretCallTimes >= 0 {
			readSecretCall.Times(readSecretCallTimes)
		}

		increaseAccessCountCall := dataStore.EXPECT().
			IncreaseAccessCount(gomock.Any(), gomock.Any()).
			Return(int64(1), nil).
			AnyTimes()

		if increaseAccessCountCallTimes >= 0 {
			increaseAccessCountCall.Times(increaseAccessCountCallTimes)
		}

		response = commands.GetSecretMetadata(context.Background(), dataStore, secret.ID)

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

		t.Run("it should return ID", func(t *testing.T) {
			assert.Equal(t, secret.ID, response.ID)
		})

		t.Run("it should return access count", func(t *testing.T) {
			assert.Equal(t, 1, response.AccessCount)
		})

		t.Run("it should return access limit", func(t *testing.T) {
			assert.Equal(t, secret.AccessLimit, response.AccessLimit)
		})

		t.Run("it should return expiration", func(t *testing.T) {
			assert.Equal(t, secret.Expiration().Format(), response.Expiration.Format())
		})
	})

	t.Run("when context is cancelled", func(t *testing.T) {
		t.Run("it should return nil", func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			cancel()

			ctrl := gomock.NewController(t)
			dataStore := mocks.NewMockDataStore(ctrl)

			response := commands.GetSecretMetadata(ctx, dataStore, secret.ID)

			assert.Nil(t, response)
		})
	})

}

func TestWhenGettingSecretMetadataForSecretThatDoesNotExist(t *testing.T) {

	sut := func(readSecretCallTimes, increaseAccessCountCallTimes int) *models.SecretMetadata {
		ctrl := gomock.NewController(t)

		dataStore := mocks.NewMockDataStore(ctrl)
		readSecretCall := dataStore.EXPECT().
			ReadSecret(gomock.Any(), gomock.Any()).
			Return(nil).
			AnyTimes()
		if readSecretCallTimes >= 0 {
			readSecretCall.Times(readSecretCallTimes)
		}

		increaseAccessCountCall := dataStore.EXPECT().
			IncreaseAccessCount(gomock.Any(), gomock.Any()).
			AnyTimes()
		if increaseAccessCountCallTimes >= 0 {
			increaseAccessCountCall.Times(increaseAccessCountCallTimes)
		}

		return commands.GetSecretMetadata(context.Background(), dataStore, testhelpers.RandomId(t))
	}

	t.Run("it should return nil", func(t *testing.T) {
		response := sut(-1, -1)
		assert.Nil(t, response)
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
			DeleteSecret(gomock.Any(), secret.ID).
			Return(true, nil).
			AnyTimes()
		if deleteSecretCallTimes >= 0 {
			deleteSecretCall.Times(deleteSecretCallTimes)
		}

		return commands.DeleteSecret(context.Background(), dataStore, secret.ID)
	}

	t.Run("should return", func(t *testing.T) {
		response, err := sut(-1)

		t.Run("it should not return error", func(t *testing.T) {
			assert.NoError(t, err)
		})

		t.Run("it should return true", func(t *testing.T) {
			assert.True(t, response)
		})
	})
	t.Run("should delete from database", func(t *testing.T) { _, _ = sut(1) })

	t.Run("when context is cancelled", func(t *testing.T) {
		t.Run("it should return context error", func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			cancel()

			ctrl := gomock.NewController(t)
			dataStore := mocks.NewMockDataStore(ctrl)

			secret := &models.Secret{
				ID: testhelpers.RandomId(t),
			}

			_, err := commands.DeleteSecret(ctx, dataStore, secret.ID)

			assert.True(t, pkgerrors.IsContextError(err), "expected context error")
		})
	})
}

func TestWhenDeletingASecretThatDoesNotExist(t *testing.T) {
	sut := func(deleteSecretCallTimes int) (response bool, err error) {

		ctrl := gomock.NewController(t)

		dataStore := mocks.NewMockDataStore(ctrl)
		deleteSecretCall := dataStore.EXPECT().
			DeleteSecret(gomock.Any(), gomock.Any()).
			Return(false, nil).
			AnyTimes()
		if deleteSecretCallTimes >= 0 {
			deleteSecretCall.Times(deleteSecretCallTimes)
		}

		id := testhelpers.RandomId(t)

		return commands.DeleteSecret(context.Background(), dataStore, id)
	}

	t.Run("should return", func(t *testing.T) {
		response, err := sut(-1)

		t.Run("it should not return error", func(t *testing.T) {
			assert.NoError(t, err)
		})

		t.Run("it should return false", func(t *testing.T) {
			assert.False(t, response)
		})
	})
	t.Run("should delete from database", func(t *testing.T) { _, _ = sut(1) })
}
