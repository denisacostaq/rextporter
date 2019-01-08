package memconfig

import (
	"testing"

	"github.com/simelo/rextporter/src/config"
	"github.com/simelo/rextporter/src/config/mocks"
	"github.com/stretchr/testify/suite"
)

func newMetricDef(suite *metricDefConfSuit) config.RextMetricDef {
	return NewMetricDef(
		suite.metricName,
		suite.metricType,
		suite.metricDescription,
		suite.nodeSolver,
		suite.metricOptions,
		suite.metricLabels,
	)
}

type metricDefConfSuit struct {
	suite.Suite
	metricDef                                 config.RextMetricDef
	metricName, metricType, metricDescription string
	nodeSolver                                config.RextNodeSolver
	metricLabels                              []config.RextLabelDef
	metricOptions                             config.RextKeyValueStore
}

func (suite *metricDefConfSuit) SetupTest() {
	suite.metricName = "MySupperMetric"
	suite.metricType = config.KeyMetricTypeCounter
	suite.metricDescription = "This is all about ..."
	suite.nodeSolver = &NodeSolver{nodePath: "sds"}
	suite.nodeSolver.GetOptions()
	suite.metricLabels = nil
	suite.metricOptions = NewOptionsMap()
	_, err := suite.metricOptions.SetString("k1", "v1")
	suite.Nil(err)
	_, err = suite.metricOptions.SetString("k2", "v2")
	suite.Nil(err)
	suite.metricDef = newMetricDef(suite)
}

func TestMetricDefConfSuit(t *testing.T) {
	suite.Run(t, new(metricDefConfSuit))
}

func (suite *metricDefConfSuit) TestNewMetricDef() {
	// NOTE(denisacostaq@gmail.com): Giving

	// NOTE(denisacostaq@gmail.com): When
	metricDef := newMetricDef(suite)
	opts, err := suite.metricOptions.Clone()
	suite.Nil(err)
	_, err = suite.metricOptions.SetString("k1", "v2")
	suite.Nil(err)

	// NOTE(denisacostaq@gmail.com): Assert
	suite.Equal(suite.metricName, metricDef.GetMetricName())
	suite.Equal(suite.metricType, metricDef.GetMetricType())
	suite.Equal(suite.metricDescription, metricDef.GetMetricDescription())
	suite.Equal(suite.nodeSolver, metricDef.GetNodeSolver())
	suite.Equal(suite.metricOptions, metricDef.GetOptions())
	suite.NotEqual(opts, metricDef.GetOptions())
	suite.Equal(suite.metricLabels, metricDef.GetLabels())
}

func (suite *metricDefConfSuit) TestAbleToSetName() {
	// NOTE(denisacostaq@gmail.com): Giving
	orgName := suite.metricDef.GetMetricName()
	name := "fgfg78"
	suite.metricDef.SetMetricName(name)

	// NOTE(denisacostaq@gmail.com): When
	name2 := suite.metricDef.GetMetricName()

	// NOTE(denisacostaq@gmail.com): Assert
	suite.Equal(name, name2)
	suite.NotEqual(orgName, name2)
}

func (suite *metricDefConfSuit) TestAbleToSetType() {
	// NOTE(denisacostaq@gmail.com): Giving
	orgT := suite.metricDef.GetMetricType()
	tp := "fgfg78"
	suite.metricDef.SetMetricType(tp)

	// NOTE(denisacostaq@gmail.com): When
	tp2 := suite.metricDef.GetMetricType()

	// NOTE(denisacostaq@gmail.com): Assert
	suite.Equal(tp, tp2)
	suite.NotEqual(orgT, tp2)
}

func (suite *metricDefConfSuit) TestAbleToSetDescription() {
	// NOTE(denisacostaq@gmail.com): Giving
	orgDescription := suite.metricDef.GetMetricDescription()
	description := "fgfg78"
	suite.metricDef.SetMetricDescription(description)

	// NOTE(denisacostaq@gmail.com): When
	description2 := suite.metricDef.GetMetricDescription()

	// NOTE(denisacostaq@gmail.com): Assert
	suite.Equal(description, description2)
	suite.NotEqual(orgDescription, description2)
}

func (suite *metricDefConfSuit) TestAbleToSetNodeSolver() {
	// NOTE(denisacostaq@gmail.com): Giving
	orgNs := suite.metricDef.GetMetricName()
	ns := &NodeSolver{nodePath: "dfdfd"}
	suite.metricDef.SetNodeSolver(ns)

	// NOTE(denisacostaq@gmail.com): When
	ns2 := suite.metricDef.GetNodeSolver()

	// NOTE(denisacostaq@gmail.com): Assert
	suite.Equal(ns, ns2)
	suite.NotEqual(orgNs, ns2)
}
func (suite *metricDefConfSuit) TestAbleToAddLabel() {
	// NOTE(denisacostaq@gmail.com): Giving
	orgLabels := suite.metricDef.GetLabels()

	// NOTE(denisacostaq@gmail.com): When
	suite.metricDef.AddLabel(&LabelDef{})
	labels2 := suite.metricDef.GetLabels()

	// NOTE(denisacostaq@gmail.com): Assert
	suite.Equal(len(orgLabels)+1, len(labels2))
}

func (suite *metricDefConfSuit) TestInitializeEmptyOptionsInFly() {
	// NOTE(denisacostaq@gmail.com): Giving

	// NOTE(denisacostaq@gmail.com): When
	metricDef := MetricDef{}

	// NOTE(denisacostaq@gmail.com): Assert
	suite.NotNil(metricDef.GetOptions())
}

func (suite *metricDefConfSuit) TestValidationClonedShouldBeValid() {
	// NOTE(denisacostaq@gmail.com): Giving

	// NOTE(denisacostaq@gmail.com): When
	cMetricDef, err := suite.metricDef.Clone()
	suite.Nil(err)
	suite.Equal(suite.metricDef, cMetricDef)
	setUpFakeValidationOn3rdPartyOverMetric(cMetricDef)
	hasError := cMetricDef.Validate()

	// NOTE(denisacostaq@gmail.com): Assert
	suite.False(hasError)
}

func (suite *metricDefConfSuit) TestValidationNameShouldNotBeEmpty() {
	// NOTE(denisacostaq@gmail.com): Giving
	cMetricDef, err := suite.metricDef.Clone()
	suite.Nil(err)
	setUpFakeValidationOn3rdPartyOverMetric(cMetricDef)

	// NOTE(denisacostaq@gmail.com): When
	cMetricDef.SetMetricName("")
	hasError := cMetricDef.Validate()

	// NOTE(denisacostaq@gmail.com): Assert
	suite.True(hasError)
}

func (suite *metricDefConfSuit) TestValidationTypeShouldNotBeEmpty() {
	// NOTE(denisacostaq@gmail.com): Giving
	cMetricDef, err := suite.metricDef.Clone()
	suite.Nil(err)
	setUpFakeValidationOn3rdPartyOverMetric(cMetricDef)

	// NOTE(denisacostaq@gmail.com): When
	cMetricDef.SetMetricType("")
	hasError := cMetricDef.Validate()

	// NOTE(denisacostaq@gmail.com): Assert
	suite.True(hasError)
}

func (suite *metricDefConfSuit) TestValidationTypeShouldBeValid() {
	// NOTE(denisacostaq@gmail.com): Giving
	cMetricDef, err := suite.metricDef.Clone()
	suite.Nil(err)
	setUpFakeValidationOn3rdPartyOverMetric(cMetricDef)

	// NOTE(denisacostaq@gmail.com): When
	cMetricDef.SetMetricType("fgfgfg")
	hasError := cMetricDef.Validate()

	// NOTE(denisacostaq@gmail.com): Assert
	suite.True(hasError)
}

func (suite *metricDefConfSuit) TestValidationNodeSolverShouldNotBeNil() {
	// NOTE(denisacostaq@gmail.com): Giving
	cMetricDef, err := suite.metricDef.Clone()
	suite.Nil(err)
	setUpFakeValidationOn3rdPartyOverMetric(cMetricDef)

	// NOTE(denisacostaq@gmail.com): When
	cMetricDef.SetNodeSolver(nil)
	hasError := cMetricDef.Validate()

	// NOTE(denisacostaq@gmail.com): Assert
	suite.True(hasError)
}

func (suite *metricDefConfSuit) TestValidationShouldGoDownTroughFields() {
	// NOTE(denisacostaq@gmail.com): Giving
	cMetricDef, err := suite.metricDef.Clone()
	suite.Nil(err)
	mockNodeSolver := new(mocks.RextNodeSolver)
	mockNodeSolver.On("Validate").Return(false)
	cMetricDef.SetNodeSolver(mockNodeSolver)
	mockLabel1 := new(mocks.RextLabelDef)
	mockLabel1.On("Validate").Return(false)
	cMetricDef.AddLabel(mockLabel1)
	mockLabel2 := new(mocks.RextLabelDef)
	mockLabel2.On("Validate").Return(false)
	cMetricDef.AddLabel(mockLabel2)

	// NOTE(denisacostaq@gmail.com): When
	cMetricDef.Validate()

	// NOTE(denisacostaq@gmail.com): Assert
	mockNodeSolver.AssertCalled(suite.T(), "Validate")
	mockLabel1.AssertCalled(suite.T(), "Validate")
	mockLabel2.AssertCalled(suite.T(), "Validate")
}

func setUpFakeValidationOn3rdPartyOverMetric(metricDef config.RextMetricDef) {
	nodeSolverStub := new(mocks.RextNodeSolver)
	nodeSolverStub.On("Validate").Return(false)
	labelStub1 := new(mocks.RextLabelDef)
	labelStub1.On("Validate").Return(false)
	labelStub2 := new(mocks.RextLabelDef)
	labelStub2.On("Validate").Return(false)
	metricDef.SetNodeSolver(nodeSolverStub)
	metricDef.AddLabel(labelStub1)
	metricDef.AddLabel(labelStub2)
}
