package memconfig

import (
	"testing"

	"github.com/simelo/rextporter/src/config"
	"github.com/simelo/rextporter/src/config/mocks"
	"github.com/stretchr/testify/suite"
)

func newRootConfig(suite *rootConfigSuite) config.RextRoot {
	return NewRootConfig(suite.services)
}

type rootConfigSuite struct {
	suite.Suite
	services   []config.RextServiceDef
	rootConfig config.RextRoot
}

func (suite *rootConfigSuite) SetupTest() {
	suite.services = nil
	suite.rootConfig = newRootConfig(suite)
}

func TestRootConfig(t *testing.T) {
	suite.Run(t, new(rootConfigSuite))
}

func (suite *rootConfigSuite) TestNewRootConf() {
	// NOTE(denisacostaq@gmail.com): Giving

	// NOTE(denisacostaq@gmail.com): When
	rootConfig := newRootConfig(suite)
	services := rootConfig.GetServices()

	// NOTE(denisacostaq@gmail.com): Assert
	suite.Equal(len(suite.services), len(services))
}

func (suite *rootConfigSuite) TestAbleToAddService() {
	// NOTE(denisacostaq@gmail.com): Giving
	orgServices := suite.rootConfig.GetServices()
	service := &Service{}
	suite.rootConfig.AddService(service)

	// NOTE(denisacostaq@gmail.com): When
	services2 := suite.rootConfig.GetServices()

	// NOTE(denisacostaq@gmail.com): Assert
	suite.Equal(len(orgServices)+1, len(services2))
}

func (suite *rootConfigSuite) TestValidationClonedShouldBeValid() {
	// NOTE(denisacostaq@gmail.com): Giving

	// NOTE(denisacostaq@gmail.com): When
	cRootConfig, err := suite.rootConfig.Clone()
	suite.Nil(err)
	suite.Equal(suite.rootConfig, cRootConfig)
	setUpFakeValidationOn3rdPartyOverRootConfig(cRootConfig)
	hasError := cRootConfig.Validate()

	// NOTE(denisacostaq@gmail.com): Assert
	suite.False(hasError)
}

func (suite *rootConfigSuite) TestValidationShouldGoDownTroughFields() {
	// NOTE(denisacostaq@gmail.com): Giving
	cRootConfig, err := suite.rootConfig.Clone()
	suite.Nil(err)
	mockService1 := new(mocks.RextServiceDef)
	mockService1.On("Validate").Return(false)
	mockService2 := new(mocks.RextServiceDef)
	mockService2.On("Validate").Return(false)
	cRootConfig.AddService(mockService1)
	cRootConfig.AddService(mockService2)

	// NOTE(denisacostaq@gmail.com): When
	cRootConfig.Validate()

	// NOTE(denisacostaq@gmail.com): Assert
	suite.Len(cRootConfig.GetServices(), 2)
	mockService2.AssertCalled(suite.T(), "Validate")
	mockService2.AssertCalled(suite.T(), "Validate")
}

func setUpFakeValidationOn3rdPartyOverRootConfig(root config.RextRoot) {
	serviceStub := new(mocks.RextServiceDef)
	serviceStub.On("Validate").Return(false)
	root.AddService(serviceStub)
}
