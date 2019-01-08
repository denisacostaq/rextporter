package memconfig

import (
	"testing"

	"github.com/simelo/rextporter/src/config"
	"github.com/stretchr/testify/suite"
)

func newNodeSolver(suite *nodeSolverSuit) config.RextNodeSolver {
	return NewNodeSolver(
		suite.nodeSolverType,
		suite.nodePath,
		suite.options,
	)
}

type nodeSolverSuit struct {
	suite.Suite
	nodeSolver               config.RextNodeSolver
	nodeSolverType, nodePath string
	options                  config.RextKeyValueStore
}

func (suite *nodeSolverSuit) SetupTest() {
	suite.nodeSolverType = config.RextNodeSolverTypeJSONPath
	suite.nodePath = "/tmp/a"
	suite.options = NewOptionsMap()
	_, err := suite.options.SetString("k1", "v1")
	suite.Nil(err)
	_, err = suite.options.SetString("k2", "v2")
	suite.Nil(err)
	suite.nodeSolver = newNodeSolver(suite)
}

func TestNodeSolverSuit(t *testing.T) {
	suite.Run(t, new(nodeSolverSuit))
}

func (suite *nodeSolverSuit) TestNewNodeSolver() {
	// NOTE(denisacostaq@gmail.com): Giving

	// NOTE(denisacostaq@gmail.com): When
	nodeSolver := newNodeSolver(suite)
	opts, err := suite.options.Clone()
	suite.Nil(err)
	_, err = suite.options.SetString("k1", "v2")
	suite.Nil(err)

	// NOTE(denisacostaq@gmail.com): Assert
	suite.Equal(suite.nodeSolverType, nodeSolver.GetType())
	suite.Equal(suite.nodePath, nodeSolver.GetNodePath())
	suite.Equal(suite.options, nodeSolver.GetOptions())
	suite.NotEqual(opts, nodeSolver.GetOptions())
}

func (suite *nodeSolverSuit) TestAbleToSetNodePath() {
	// NOTE(denisacostaq@gmail.com): Giving
	orgNodePath := suite.nodeSolver.GetNodePath()
	nodePath := "fgfg78"
	suite.nodeSolver.SetNodePath(nodePath)

	// NOTE(denisacostaq@gmail.com): When
	nodePath2 := suite.nodeSolver.GetNodePath()

	// NOTE(denisacostaq@gmail.com): Assert
	suite.Equal(nodePath, nodePath2)
	suite.NotEqual(orgNodePath, nodePath2)
}

func (suite *nodeSolverSuit) TestValidationClonedShouldBeValid() {
	// NOTE(denisacostaq@gmail.com): Giving

	// NOTE(denisacostaq@gmail.com): When
	cNodeSolver, err := suite.nodeSolver.Clone()
	suite.Nil(err)
	suite.Equal(suite.nodeSolver, cNodeSolver)
	hasError := cNodeSolver.Validate()

	// NOTE(denisacostaq@gmail.com): Assert
	suite.False(hasError)
}

func (suite *nodeSolverSuit) TestValidationTypeShouldNotBeEmpty() {
	// NOTE(denisacostaq@gmail.com): Giving
	nodeSolver, err := suite.nodeSolver.Clone()
	suite.Nil(err)

	// NOTE(denisacostaq@gmail.com): When
	nodeSolver.SetNodePath("")
	hasError := nodeSolver.Validate()

	// NOTE(denisacostaq@gmail.com): Assert
	suite.True(hasError)
}
