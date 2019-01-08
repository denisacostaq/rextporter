package memconfig

import (
	"testing"

	"github.com/simelo/rextporter/src/config"
	"github.com/stretchr/testify/suite"
)

func newAuth(suite *authConfSuit) config.RextAuthDef {
	return NewHTTPAuth(suite.authType, suite.authURL, suite.options)
}

type authConfSuit struct {
	suite.Suite
	authConf          config.RextAuthDef
	authType, authURL string
	options           OptionsMap
}

func (suite *authConfSuit) SetupTest() {
	suite.options = NewOptionsMap()
	_, err := suite.options.SetString(config.OptKeyRextAuthDefTokenHeaderKey, "v1")
	suite.Nil(err)
	_, err = suite.options.SetString(config.OptKeyRextAuthDefTokenGenEndpoint, "v2")
	suite.Nil(err)
	_, err = suite.options.SetString(config.OptKeyRextAuthDefTokenKeyFromEndpoint, "v3")
	suite.Nil(err)
	suite.authType = "CSRF"
	suite.authURL = "http://localhost:9000/hosted_in/auth"
	suite.authConf = newAuth(suite)
}

func TestAuthConfSuit(t *testing.T) {
	suite.Run(t, new(authConfSuit))
}

func (suite *authConfSuit) TestNewHTTPAuth() {
	// NOTE(denisacostaq@gmail.com): Giving

	// NOTE(denisacostaq@gmail.com): When
	authConf := newAuth(suite)
	opts, err := suite.options.Clone()
	suite.Nil(err)
	_, err = suite.options.SetString("k1", "v2")
	suite.Nil(err)

	// NOTE(denisacostaq@gmail.com): Assert
	suite.Equal(suite.authType, authConf.GetAuthType())
	suite.Equal(suite.options, authConf.GetOptions())
	suite.NotEqual(opts, authConf.GetOptions())
}

func (suite *authConfSuit) TestAbleToSetType() {
	// NOTE(denisacostaq@gmail.com): Giving
	orgAuthType := suite.authConf.GetAuthType()
	authType := "fgfg78"
	suite.authConf.SetAuthType(authType)

	// NOTE(denisacostaq@gmail.com): When
	authType2 := suite.authConf.GetAuthType()

	// NOTE(denisacostaq@gmail.com): Assert
	suite.Equal(authType, authType2)
	suite.NotEqual(orgAuthType, authType2)
}

func (suite *authConfSuit) TestInitializeEmptyOptionsInFly() {
	// NOTE(denisacostaq@gmail.com): Giving

	// NOTE(denisacostaq@gmail.com): When
	authDef := HTTPAuth{}

	// NOTE(denisacostaq@gmail.com): Assert
	suite.NotNil(authDef.GetOptions())
}

func (suite *authConfSuit) TestValidationClonedShouldBeValid() {
	// NOTE(denisacostaq@gmail.com): Giving

	// NOTE(denisacostaq@gmail.com): When
	cAuthConf, err := suite.authConf.Clone()
	suite.Nil(err)
	suite.Equal(suite.authConf, cAuthConf)
	hasError := cAuthConf.Validate()

	// NOTE(denisacostaq@gmail.com): Assert
	suite.False(hasError)
}

func (suite *authConfSuit) TestValidationTypeShouldNotBeEmpty() {
	// NOTE(denisacostaq@gmail.com): Giving
	authDef, err := suite.authConf.Clone()
	suite.Nil(err)

	// NOTE(denisacostaq@gmail.com): When
	authDef.SetAuthType("")
	hasError := authDef.Validate()

	// NOTE(denisacostaq@gmail.com): Assert
	suite.True(hasError)
}

func (suite *authConfSuit) TestValidationTokenHeaderKeyShouldNotBeEmptyInCSRF() {
	// NOTE(denisacostaq@gmail.com): Giving
	authDef, err := suite.authConf.Clone()
	suite.Nil(err)

	// NOTE(denisacostaq@gmail.com): When
	opts := authDef.GetOptions()
	pe, err := opts.SetString(config.OptKeyRextAuthDefTokenHeaderKey, "")
	suite.True(pe)
	suite.Nil(err)
	hasError := authDef.Validate()

	// NOTE(denisacostaq@gmail.com): Assert
	suite.True(hasError)
}

func (suite *authConfSuit) TestValidationTokenGenEndpointShouldNotBeEmptyInCSRF() {
	// NOTE(denisacostaq@gmail.com): Giving
	authDef, err := suite.authConf.Clone()
	suite.Nil(err)

	// NOTE(denisacostaq@gmail.com): When
	opts := authDef.GetOptions()
	pe, err := opts.SetString(config.OptKeyRextAuthDefTokenGenEndpoint, "")
	suite.True(pe)
	suite.Nil(err)
	hasError := authDef.Validate()

	// NOTE(denisacostaq@gmail.com): Assert
	suite.True(hasError)
}

func (suite *authConfSuit) TestValidationTokenKeyFromEndpointShouldNotBeEmptyInCSRF() {
	// NOTE(denisacostaq@gmail.com): Giving
	authDef, err := suite.authConf.Clone()
	suite.Nil(err)

	// NOTE(denisacostaq@gmail.com): When
	opts := authDef.GetOptions()
	pe, err := opts.SetString(config.OptKeyRextAuthDefTokenKeyFromEndpoint, "")
	suite.True(pe)
	suite.Nil(err)
	hasError := authDef.Validate()

	// NOTE(denisacostaq@gmail.com): Assert
	suite.True(hasError)
}

func (suite *authConfSuit) TestValidationTokenValsCanBeEmptyInNotCSRF() {
	// NOTE(denisacostaq@gmail.com): Giving
	authDef, err := suite.authConf.Clone()
	suite.Nil(err)
	authDef.SetAuthType("tt3")

	// NOTE(denisacostaq@gmail.com): When
	opts := authDef.GetOptions()
	pe, err := opts.SetString(config.OptKeyRextAuthDefTokenKeyFromEndpoint, "")
	suite.True(pe)
	suite.Nil(err)
	pe, err = opts.SetString(config.OptKeyRextAuthDefTokenGenEndpoint, "")
	suite.True(pe)
	suite.Nil(err)
	pe, err = opts.SetString(config.OptKeyRextAuthDefTokenHeaderKey, "")
	suite.True(pe)
	suite.Nil(err)
	hasError := authDef.Validate()

	// NOTE(denisacostaq@gmail.com): Assert
	suite.False(hasError)
}
