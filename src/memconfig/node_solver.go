package memconfig

import (
	"github.com/simelo/rextporter/src/config"
	log "github.com/sirupsen/logrus"
)

// NodeSolver implements the interface config.RextNodeSolver
type NodeSolver struct {
	// FIXME(denisacostaq@gmail.com): lowercase
	MType    string
	nodePath string
	options  config.RextKeyValueStore
}

// Clone make a deep copy of NodeSolver or return an error if any
func (ns NodeSolver) Clone() (cNs config.RextNodeSolver, err error) {
	var cOpts config.RextKeyValueStore
	if cOpts, err = ns.GetOptions().Clone(); err != nil {
		log.WithError(err).Errorln("can not clone options in node solver")
		return cNs, err
	}
	cNs = NewNodeSolver(ns.MType, ns.nodePath, cOpts)
	return cNs, err
}

// GetType return solver type
func (ns NodeSolver) GetType() string {
	return ns.MType
}

// SetNodePath set the node path
func (ns *NodeSolver) SetNodePath(nodePath string) {
	ns.nodePath = nodePath
}

// GetNodePath return the node path
func (ns NodeSolver) GetNodePath() string {
	return ns.nodePath
}

// GetOptions return key/value pairs for extra options
func (ns *NodeSolver) GetOptions() config.RextKeyValueStore {
	if ns.options == nil {
		ns.options = NewOptionsMap()
	}
	return ns.options
}

// Validate the node solver, return true if any error is found
func (ns NodeSolver) Validate() bool {
	return config.ValidateNodeSolver(&ns)
}

// NewNodeSolver create a new node solver
func NewNodeSolver(mType, nodePath string, options config.RextKeyValueStore) config.RextNodeSolver {
	return &NodeSolver{
		MType:    mType,
		nodePath: nodePath,
		options:  options,
	}
}
