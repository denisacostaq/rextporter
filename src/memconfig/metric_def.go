package memconfig

import (
	"github.com/simelo/rextporter/src/config"
	log "github.com/sirupsen/logrus"
)

// MetricDef implements the interface config.RextMetricDef
type MetricDef struct {
	name        string
	mType       string
	nodeSolver  config.RextNodeSolver
	description string
	labels      []config.RextLabelDef
	options     config.RextKeyValueStore
}

// Clone make a deep copy of MetricDef or return an error if any
func (m MetricDef) Clone() (cM config.RextMetricDef, err error) {
	var cNs config.RextNodeSolver
	if m.GetNodeSolver() != nil {
		if cNs, err = m.GetNodeSolver().Clone(); err != nil {
			log.WithError(err).Errorln("can not clone node solver in metric")
			return cM, err
		}
	}
	var cLabels []config.RextLabelDef
	for _, label := range m.labels {
		var cLabel config.RextLabelDef
		if cLabel, err = label.Clone(); err != nil {
			log.WithError(err).Errorln("can not clone labels in metric")
			return cM, err
		}
		cLabels = append(cLabels, cLabel)
	}
	var cOpts config.RextKeyValueStore
	if cOpts, err = m.GetOptions().Clone(); err != nil {
		log.WithError(err).Errorln("can not clone options in metric")
		return cM, err
	}
	cM = NewMetricDef(m.GetMetricName(), m.GetMetricType(), m.GetMetricDescription(), cNs, cOpts, cLabels)
	return cM, err
}

// GetMetricName return the metric name
func (m MetricDef) GetMetricName() string {
	return m.name
}

// GetMetricType return the metric type, one of: gauge, counter, histogram or summary
func (m MetricDef) GetMetricType() string {
	return m.mType
}

// GetNodeSolver return solver type
func (m MetricDef) GetNodeSolver() config.RextNodeSolver {
	return m.nodeSolver
}

// SetNodeSolver set the node solver
func (m *MetricDef) SetNodeSolver(nodeSolver config.RextNodeSolver) {
	m.nodeSolver = nodeSolver
}

// GetMetricDescription return the metric description
func (m MetricDef) GetMetricDescription() string {
	return m.description
}

// SetMetricName can set the metric name
func (m *MetricDef) SetMetricName(name string) {
	m.name = name
}

// SetMetricType can set the metric type
func (m *MetricDef) SetMetricType(tp string) {
	m.mType = tp
}

// SetMetricDescription can set the metric description
func (m *MetricDef) SetMetricDescription(description string) {
	m.description = description
}

// GetLabels return labels
func (m MetricDef) GetLabels() []config.RextLabelDef {
	return m.labels
}

// AddLabel receive label to be append to the current list
func (m *MetricDef) AddLabel(label config.RextLabelDef) {
	m.labels = append(m.labels, label)
}

// GetOptions return key/value pairs for extra options
func (m *MetricDef) GetOptions() config.RextKeyValueStore {
	if m.options == nil {
		m.options = NewOptionsMap()
	}
	return m.options
}

// Validate the metric, return true if any error is found
func (m MetricDef) Validate() bool {
	return config.ValidateMetric(&m)
}

// NewMetricDef create a new metric definition
func NewMetricDef(name, mType, description string, nodeSolver config.RextNodeSolver, options config.RextKeyValueStore, labels []config.RextLabelDef) *MetricDef {
	return &MetricDef{
		name:        name,
		mType:       mType,
		nodeSolver:  nodeSolver,
		description: description,
		labels:      labels,
		options:     options,
	}
}
