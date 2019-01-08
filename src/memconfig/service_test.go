package memconfig

import (
	"testing"

	"github.com/simelo/rextporter/src/config"
	"github.com/simelo/rextporter/src/config/mocks"
	"github.com/stretchr/testify/suite"
)

type serviceConfSuit struct {
	suite.Suite
	srvConf   config.RextServiceDef
	basePath  string
	protocol  string
	auth      config.RextAuthDef
	resources []config.RextResourceDef
	options   config.RextKeyValueStore
}

func newService(suite *serviceConfSuit) config.RextServiceDef {
	return NewServiceConf(suite.basePath, suite.protocol, suite.auth, suite.resources, suite.options)
}

func (suite *serviceConfSuit) SetupTest() {
	suite.basePath = "/hosted_in/root"
	suite.protocol = "file"
	suite.auth = &HTTPAuth{}
	suite.auth.GetOptions()
	suite.resources = nil
	suite.options = NewOptionsMap()
	_, err := suite.options.SetString(config.OptKeyRextServiceDefJobName, "v1")
	suite.Nil(err)
	_, err = suite.options.SetString(config.OptKeyRextServiceDefInstanceName, "v2")
	suite.Nil(err)
	suite.srvConf = newService(suite)
}

func TestServiceConfSuit(t *testing.T) {
	suite.Run(t, new(serviceConfSuit))
}

func (suite *serviceConfSuit) TestNewServiceDef() {
	// NOTE(denisacostaq@gmail.com): Giving

	// NOTE(denisacostaq@gmail.com): When
	serviceDef := newService(suite)
	opts, err := suite.options.Clone()
	suite.Nil(err)
	_, err = suite.options.SetString("k1", "v2")
	suite.Nil(err)

	// NOTE(denisacostaq@gmail.com): Assert
	suite.Equal(suite.protocol, serviceDef.GetProtocol())
	suite.Equal(suite.options, serviceDef.GetOptions())
	suite.NotEqual(opts, serviceDef.GetOptions())
}

func (suite *serviceConfSuit) TestAbleToSetProtocol() {
	// NOTE(denisacostaq@gmail.com): Giving
	protocol := "http"
	suite.srvConf.SetProtocol(protocol)

	// NOTE(denisacostaq@gmail.com): When
	protocol2 := suite.srvConf.GetProtocol()

	// NOTE(denisacostaq@gmail.com): Assert
	suite.Equal(protocol, protocol2)
}

func (suite *serviceConfSuit) TestAbleToSetBasePath() {
	// NOTE(denisacostaq@gmail.com): Giving
	basePath := "dfdf"
	suite.srvConf.SetBasePath(basePath)

	// NOTE(denisacostaq@gmail.com): When
	basePath2 := suite.srvConf.GetBasePath()

	// NOTE(denisacostaq@gmail.com): Assert
	suite.Equal(basePath, basePath2)
}

func (suite *serviceConfSuit) TestAbleToSetBaseAuth() {
	// NOTE(denisacostaq@gmail.com): Giving
	orgAuth := suite.srvConf.GetAuthForBaseURL()
	auth := &HTTPAuth{authType: "ds"}
	suite.srvConf.SetAuthForBaseURL(auth)

	// NOTE(denisacostaq@gmail.com): When
	auth2 := suite.srvConf.GetAuthForBaseURL()

	// NOTE(denisacostaq@gmail.com): Assert
	suite.Equal(auth, auth2)
	suite.NotEqual(orgAuth, auth2)
}

func (suite *serviceConfSuit) TestAbleToAddSource() {
	// NOTE(denisacostaq@gmail.com): Giving
	orgResource := suite.srvConf.GetResources()
	resource := &ResourceDef{}
	suite.srvConf.AddResource(resource)

	// NOTE(denisacostaq@gmail.com): When
	resource2 := suite.srvConf.GetResources()

	// NOTE(denisacostaq@gmail.com): Assert
	suite.Equal(len(orgResource)+1, len(resource2))
}

func (suite *serviceConfSuit) TestInitializeEmptyOptionsInFly() {
	// NOTE(denisacostaq@gmail.com): Giving

	// NOTE(denisacostaq@gmail.com): When
	srvConf := Service{}

	// NOTE(denisacostaq@gmail.com): Assert
	suite.NotNil(srvConf.GetOptions())
}

func (suite *serviceConfSuit) TestValidationClonedShouldBeValid() {
	// NOTE(denisacostaq@gmail.com): Giving

	// NOTE(denisacostaq@gmail.com): When
	cSrvConf, err := suite.srvConf.Clone()
	suite.Nil(err)
	suite.Equal(suite.srvConf, cSrvConf)
	setUpFakeValidationOn3rdPartyOverService(cSrvConf)
	hasError := cSrvConf.Validate()

	// NOTE(denisacostaq@gmail.com): Assert
	suite.False(hasError)
}

func (suite *serviceConfSuit) TestValidationJobNameShouldNotBeEmpty() {
	// NOTE(denisacostaq@gmail.com): Giving
	srvConf, err := suite.srvConf.Clone()
	suite.Nil(err)

	// NOTE(denisacostaq@gmail.com): When
	opts := srvConf.GetOptions()
	var pe bool
	pe, err = opts.SetString(config.OptKeyRextServiceDefJobName, "")
	suite.True(pe)
	suite.Nil(err)
	hasError := srvConf.Validate()

	// NOTE(denisacostaq@gmail.com): Assert
	suite.True(hasError)
}

func (suite *serviceConfSuit) TestValidationInstanceNameShouldNotBeEmpty() {
	// NOTE(denisacostaq@gmail.com): Giving
	srvConf, err := suite.srvConf.Clone()
	suite.Nil(err)

	// NOTE(denisacostaq@gmail.com): When
	opts := srvConf.GetOptions()
	pe, err := opts.SetString(config.OptKeyRextServiceDefInstanceName, "")
	suite.True(pe)
	suite.Nil(err)
	hasError := srvConf.Validate()

	// NOTE(denisacostaq@gmail.com): Assert
	suite.True(hasError)
}

func (suite *serviceConfSuit) TestValidationEmptyProtocol() {
	// NOTE(denisacostaq@gmail.com): Giving

	// NOTE(denisacostaq@gmail.com): When
	cSrvConf, err := suite.srvConf.Clone()
	suite.Nil(err)
	cSrvConf.SetProtocol("")
	hasError := cSrvConf.Validate()

	// NOTE(denisacostaq@gmail.com): Assert
	suite.True(hasError)
}

func (suite *serviceConfSuit) TestValidationShouldGoDownTroughFields() {
	// NOTE(denisacostaq@gmail.com): Giving
	cSrvConf, err := suite.srvConf.Clone()
	suite.Nil(err)
	mockAuth := new(mocks.RextAuthDef)
	mockAuth.On("Validate").Return(false)
	mockResource1 := new(mocks.RextResourceDef)
	mockResource1.On("Validate").Return(false)
	mockResource2 := new(mocks.RextResourceDef)
	mockResource2.On("Validate").Return(false)
	cSrvConf.SetAuthForBaseURL(mockAuth)
	cSrvConf.AddResources(mockResource1, mockResource2)

	// NOTE(denisacostaq@gmail.com): When
	cSrvConf.Validate()

	// NOTE(denisacostaq@gmail.com): Assert
	mockAuth.AssertCalled(suite.T(), "Validate")
	suite.Len(cSrvConf.GetResources(), 2)
	mockResource1.AssertCalled(suite.T(), "Validate")
	mockResource2.AssertCalled(suite.T(), "Validate")
}

func setUpFakeValidationOn3rdPartyOverService(srv config.RextServiceDef) {
	authStub := new(mocks.RextAuthDef)
	authStub.On("Validate").Return(false)
	resourceStub := new(mocks.RextResourceDef)
	resourceStub.On("Validate").Return(false)
	srv.SetAuthForBaseURL(authStub)
	srv.AddResource(resourceStub)
}
