package memconfig

import (
	"github.com/simelo/rextporter/src/config"
	log "github.com/sirupsen/logrus"
)

// LabelDef implements the interface config.RextLabelDef
type LabelDef struct {
	name       string
	nodeSolver config.RextNodeSolver
}

// Clone make a deep copy of LabelDef or return an error if any
func (l LabelDef) Clone() (cL config.RextLabelDef, err error) {
	var cNs config.RextNodeSolver
	if l.GetNodeSolver() != nil {
		if cNs, err = l.GetNodeSolver().Clone(); err != nil {
			log.WithError(err).Errorln("can not clone node solver in label")
			return cL, err
		}
	}
	cL = NewLabelDef(l.name, cNs)
	return cL, err
}

// GetName return the label name
func (l LabelDef) GetName() string {
	return l.name
}

// SetName set the name for the label
func (l *LabelDef) SetName(name string) {
	l.name = name
}

// GetNodeSolver return the solver for the label value
func (l LabelDef) GetNodeSolver() config.RextNodeSolver {
	return l.nodeSolver
}

// SetNodeSolver set the solver for the label value
func (l *LabelDef) SetNodeSolver(nodeSolver config.RextNodeSolver) {
	l.nodeSolver = nodeSolver
}

// Validate the label, return true if any error is found
func (l LabelDef) Validate() bool {
	return config.ValidateLabel(&l)
}

// NewLabelDef create a new label definition
func NewLabelDef(name string, nodeSolver config.RextNodeSolver) *LabelDef {
	return &LabelDef{
		name:       name,
		nodeSolver: nodeSolver,
	}
}
