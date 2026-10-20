package middleware

import (
	"cellar/pkg/mocks"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestSelectDatastore(t *testing.T) {
	testCases := []struct {
		description   string
		redisEnabled  bool
		valkeyEnabled bool
		expectedKind  datastoreKind
		expectedError string
	}{
		{description: "no datastore is enabled", expectedError: "at least one datastore is required"},
		{description: "only redis is enabled", redisEnabled: true, expectedKind: redisDatastore},
		{description: "only valkey is enabled", valkeyEnabled: true, expectedKind: valkeyDatastore},
		{description: "both datastores are enabled", redisEnabled: true, valkeyEnabled: true, expectedError: "cannot enable more than one datastore"},
	}

	for _, tc := range testCases {
		t.Run("when "+tc.description, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			redisCfg := mocks.NewMockIRedisConfiguration(ctrl)
			redisCfg.EXPECT().Enabled().Return(tc.redisEnabled).AnyTimes()
			valkeyCfg := mocks.NewMockIValkeyConfiguration(ctrl)
			valkeyCfg.EXPECT().Enabled().Return(tc.valkeyEnabled).AnyTimes()
			cfg := mocks.NewMockIDatastoreConfiguration(ctrl)
			cfg.EXPECT().Redis().Return(redisCfg).AnyTimes()
			cfg.EXPECT().Valkey().Return(valkeyCfg).AnyTimes()

			kind, err := selectDatastore(cfg)

			if tc.expectedError != "" {
				t.Run("it should fail with \""+tc.expectedError+"\"", func(t *testing.T) {
					assert.EqualError(t, err, tc.expectedError)
				})
				return
			}

			t.Run("it should select the enabled datastore", func(t *testing.T) {
				assert.Equal(t, tc.expectedKind, kind)
			})
		})
	}
}
