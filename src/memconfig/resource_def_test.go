package memconfig

import (
	"testing"

	"github.com/simelo/rextporter/src/config"
	"github.com/simelo/rextporter/src/config/mocks"
	"github.com/stretchr/testify/suite"
)

func newResourceDef(suite *resourceDefSuit) config.RextResourceDef {
	return NewResourceDef(
		suite.mType,
		suite.resourceURI,
		suite.auth,
		suite.metrics,
		suite.decoder,
		suite.options,
	)
}

type resourceDefSuit struct {
	suite.Suite
	mType       string
	resourceDef config.RextResourceDef
	resourceURI string
	auth        config.RextAuthDef
	decoder     config.RextDecoderDef
	metrics     []config.RextMetricDef
	options     config.RextKeyValueStore
}

func (suite *resourceDefSuit) SetupTest() {
	suite.auth = &HTTPAuth{}
	suite.auth.GetOptions()
	suite.mType = "tt"
	suite.resourceURI = "ddrer"
	suite.decoder = &Decoder{}
	suite.decoder.GetOptions()
	suite.metrics = nil
	suite.options = NewOptionsMap()
	_, err := suite.options.SetString("k1", "v1")
	suite.Nil(err)
	_, err = suite.options.SetString("k2", "v2")
	suite.Nil(err)
	suite.resourceDef = newResourceDef(suite)
}

func TestResourceDefSuitSuit(t *testing.T) {
	suite.Run(t, new(resourceDefSuit))
}

func (suite *resourceDefSuit) TestNewResourceDefSuit() {
	// NOTE(denisacostaq@gmail.com): Giving

	// NOTE(denisacostaq@gmail.com): When
	resourceDef := newResourceDef(suite)
	opts, err := suite.options.Clone()
	suite.Nil(err)
	_, err = suite.options.SetString("k1", "v2")
	suite.Nil(err)
	basePath := "dssds"

	// NOTE(denisacostaq@gmail.com): Assert
	suite.Equal(suite.mType, resourceDef.GetType())
	suite.Equal(suite.auth, resourceDef.GetAuth(nil))
	suite.Equal(basePath+suite.resourceURI, resourceDef.GetResourcePATH(basePath))
	suite.Equal(suite.decoder, resourceDef.GetDecoder())
	suite.Equal(suite.options, resourceDef.GetOptions())
	suite.NotEqual(opts, resourceDef.GetOptions())
}

func (suite *resourceDefSuit) TestAbleToSetDecoder() {
	// NOTE(denisacostaq@gmail.com): Giving
	orgDecoder := suite.resourceDef.GetDecoder()
	decoder := &Decoder{mType: "t1"}
	suite.resourceDef.SetDecoder(decoder)

	// NOTE(denisacostaq@gmail.com): When
	decoder2 := suite.resourceDef.GetDecoder()

	// NOTE(denisacostaq@gmail.com): Assert
	suite.Equal(decoder, decoder2)
	suite.NotEqual(orgDecoder, decoder2)
}

func (suite *resourceDefSuit) TestAbleToSetURI() {
	// NOTE(denisacostaq@gmail.com): Giving
	basePath := "fffd"
	orgPath := suite.resourceDef.GetResourcePATH(basePath)
	resourceURI := "uri1"
	suite.resourceDef.SetResourceURI(resourceURI)

	// NOTE(denisacostaq@gmail.com): When
	resourceURL := suite.resourceDef.GetResourcePATH(basePath)

	// NOTE(denisacostaq@gmail.com): Assert
	suite.Equal(basePath+resourceURI, resourceURL)
	suite.NotEqual(orgPath, resourceURL)
}

func (suite *resourceDefSuit) TestAbleToAddMetricDef() {
	// NOTE(denisacostaq@gmail.com): Giving
	orgMetrics := suite.resourceDef.GetMetricDefs()
	metric := &MetricDef{}
	suite.resourceDef.AddMetricDef(metric)

	// NOTE(denisacostaq@gmail.com): When
	metric2 := suite.resourceDef.GetMetricDefs()

	// NOTE(denisacostaq@gmail.com): Assert
	suite.Equal(len(orgMetrics)+1, len(metric2))
}

func (suite *resourceDefSuit) TestAbleToSetType() {
	// NOTE(denisacostaq@gmail.com): Giving
	orgType := suite.resourceDef.GetType()
	mType := "dgfg"
	suite.resourceDef.SetType(mType)

	// NOTE(denisacostaq@gmail.com): When
	mType2 := suite.resourceDef.GetType()

	// NOTE(denisacostaq@gmail.com): Assert
	suite.Equal(mType, mType2)
	suite.NotEqual(orgType, mType2)
}

func (suite *resourceDefSuit) TestInitializeEmptyOptionsInFly() {
	// NOTE(denisacostaq@gmail.com): Giving

	// NOTE(denisacostaq@gmail.com): When
	resDef := ResourceDef{}

	// NOTE(denisacostaq@gmail.com): Assert
	suite.NotNil(resDef.GetOptions())
}

func (suite *resourceDefSuit) TestValidationClonedShouldBeValid() {
	// NOTE(denisacostaq@gmail.com): Giving

	// NOTE(denisacostaq@gmail.com): When
	cResConf, err := suite.resourceDef.Clone()
	suite.Nil(err)
	suite.Equal(suite.resourceDef, cResConf)
	setUpFakeValidationOn3rdPartyOverResource(cResConf)
	hasError := cResConf.Validate()

	// NOTE(denisacostaq@gmail.com): Assert
	suite.False(hasError)
}

func (suite *resourceDefSuit) TestValidationEmptyType() {
	// NOTE(denisacostaq@gmail.com): Giving
	cResConf, err := suite.resourceDef.Clone()
	suite.Nil(err)
	setUpFakeValidationOn3rdPartyOverResource(cResConf)

	// NOTE(denisacostaq@gmail.com): When
	cResConf.SetType("")
	hasError := cResConf.Validate()

	// NOTE(denisacostaq@gmail.com): Assert
	suite.True(hasError)
}

func (suite *resourceDefSuit) TestValidationEmptyResourceURI() {
	// NOTE(denisacostaq@gmail.com): Giving
	cResConf, err := suite.resourceDef.Clone()
	suite.Nil(err)
	setUpFakeValidationOn3rdPartyOverResource(cResConf)

	// NOTE(denisacostaq@gmail.com): When
	cResConf.SetResourceURI("")
	hasError := cResConf.Validate()

	// NOTE(denisacostaq@gmail.com): Assert
	suite.True(hasError)
}

func (suite *resourceDefSuit) TestValidationNilDecoder() {
	// NOTE(denisacostaq@gmail.com): Giving
	cResConf, err := suite.resourceDef.Clone()
	suite.Nil(err)
	setUpFakeValidationOn3rdPartyOverResource(cResConf)

	// NOTE(denisacostaq@gmail.com): When
	cResConf.SetDecoder(nil)
	hasError := cResConf.Validate()

	// NOTE(denisacostaq@gmail.com): Assert
	suite.True(hasError)
}

func (suite *resourceDefSuit) TestValidationShouldGoDownTroughFields() {
	// NOTE(denisacostaq@gmail.com): Giving
	cResConf, err := suite.resourceDef.Clone()
	suite.Nil(err)
	mockAuth := new(mocks.RextAuthDef)
	mockAuth.On("Validate").Return(false)
	mockDecoder := new(mocks.RextDecoderDef)
	mockDecoder.On("Validate").Return(false)
	mockMetric1 := new(mocks.RextMetricDef)
	mockMetric1.On("Validate").Return(false)
	mockMetric2 := new(mocks.RextMetricDef)
	mockMetric2.On("Validate").Return(false)
	cResConf.SetAuth(mockAuth)
	cResConf.SetDecoder(mockDecoder)
	cResConf.AddMetricDef(mockMetric1)
	cResConf.AddMetricDef(mockMetric2)

	// NOTE(denisacostaq@gmail.com): When
	cResConf.Validate()

	// NOTE(denisacostaq@gmail.com): Assert
	mockDecoder.AssertCalled(suite.T(), "Validate")
	mockAuth.AssertCalled(suite.T(), "Validate")
	suite.Len(cResConf.GetMetricDefs(), 2)
	mockMetric1.AssertCalled(suite.T(), "Validate")
	mockMetric2.AssertCalled(suite.T(), "Validate")
}

func setUpFakeValidationOn3rdPartyOverResource(res config.RextResourceDef) {
	authStub := new(mocks.RextAuthDef)
	authStub.On("Validate").Return(false)
	decoderStub := new(mocks.RextDecoderDef)
	decoderStub.On("Validate").Return(false)
	metricStub := new(mocks.RextMetricDef)
	metricStub.On("Validate").Return(false)
	res.SetAuth(authStub)
	res.SetDecoder(decoderStub)
	res.AddMetricDef(metricStub)
}
