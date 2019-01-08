package config

import (
	log "github.com/sirupsen/logrus"
)

// ValidateAuth check if the auth instance in parameter fill the required constraints
// to be considered as a valid RextAuthDef.
// Return true if any error is found
func ValidateAuth(auth RextAuthDef) (hasError bool) {
	if len(auth.GetAuthType()) == 0 {
		hasError = true
		log.Errorln("type is required in auth config")
	}
	if auth.GetAuthType() == AuthTypeCSRF {
		opts := auth.GetOptions()
		if tkhk, err := opts.GetString(OptKeyRextAuthDefTokenHeaderKey); err != nil || len(tkhk) == 0 {
			hasError = true
			log.Errorln("token header key is required for CSRF auth type")
		}
		if tkge, err := opts.GetString(OptKeyRextAuthDefTokenGenEndpoint); err != nil || len(tkge) == 0 {
			hasError = true
			log.Errorln("token gen endpoint is required for CSRF auth type")
		}
		if tkfe, err := opts.GetString(OptKeyRextAuthDefTokenKeyFromEndpoint); err != nil || len(tkfe) == 0 {
			hasError = true
			log.Errorln("token from endpoint is required for CSRF auth type")
		}
	}
	return hasError
}

// ValidateResource check if the resource instance in parameter fill the required constraints
// to be considered as a valid RextResourceDef.
// Return true if any error is found
func ValidateResource(r RextResourceDef) (hasError bool) {
	if len(r.GetType()) == 0 {
		hasError = true
		log.Errorln("type is required in metric config")
	}
	if len(r.GetResourcePATH("")) == 0 {
		hasError = true
		log.Errorln("resource path is required in metric config")
	}
	if r.GetDecoder() == nil {
		hasError = true
		log.Errorln("decoder is required in metric config")
	} else if r.GetDecoder().Validate() {
		hasError = true
	}
	if r.GetAuth(nil) != nil {
		if r.GetAuth(nil).Validate() {
			hasError = true
		}
	}
	for _, mtrDef := range r.GetMetricDefs() {
		if mtrDef.Validate() {
			hasError = true
		}
	}
	return hasError
}

// ValidateService check if the resource instance in parameter fill the required constraints
// to be considered as a valid RextServiceDef.
// Return true if any error is found
func ValidateService(srv RextServiceDef) (hasError bool) {
	srvOpts := srv.GetOptions()
	jobName, err := srvOpts.GetString(OptKeyRextServiceDefJobName)
	if err != nil {
		hasError = true
		log.WithError(err).Errorln("key for job name not present service config")
	}
	if len(jobName) == 0 {
		hasError = true
		log.Errorln("job name is required in service config")
	}
	var instanceName string
	instanceName, err = srvOpts.GetString(OptKeyRextServiceDefInstanceName)
	if err != nil {
		hasError = true
		log.WithError(err).Errorln("key for job name not present service config")
	}
	if len(instanceName) == 0 {
		hasError = true
		log.Errorln("instance name is required in service config")
	}
	if len(srv.GetProtocol()) == 0 {
		hasError = true
		log.Errorln("protocol should not be null in service config")
	}
	if srv.GetAuthForBaseURL() != nil {
		if srv.GetAuthForBaseURL().Validate() {
			hasError = true
		}
	}
	for _, resource := range srv.GetResources() {
		if resource.Validate() {
			hasError = true
		}
	}
	return hasError
}

// ValidateNodeSolver check if the node solver instance in parameter fill the required constraints
// to be considered as a valid RextNodeSolver.
// Return true if any error is found
func ValidateNodeSolver(ns RextNodeSolver) (hasError bool) {
	if len(ns.GetNodePath()) == 0 {
		hasError = true
		log.Errorln("node path is required in node solver config")
	}
	return hasError
}

// ValidateLabel check if the label instance in parameter fill the required constraints
// to be considered as a valid RextLabelDef.
// Return true if any error is found
func ValidateLabel(l RextLabelDef) (hasError bool) {
	if len(l.GetName()) == 0 {
		hasError = true
		log.Errorln("name is required in label config")
	}
	if l.GetNodeSolver() == nil {
		hasError = true
		log.Errorln("node solver is required in label config")
	} else if l.GetNodeSolver().Validate() {
		hasError = true
	}
	return hasError
}

// ValidateDecoder check if the decoder instance in parameter fill the required constraints
// to be considered as a valid RextDecoderDef.
// Return true if any error is found
func ValidateDecoder(d RextDecoderDef) (hasError bool) {
	if len(d.GetType()) == 0 {
		hasError = true
		log.Errorln("type is required in decoder config")
	}
	return hasError
}

// ValidateMetric check if the metric instance in parameter fill the required constraints
// to be considered as a valid RextMetricDef.
// Return true if any error is found
func ValidateMetric(m RextMetricDef) (hasError bool) {
	if len(m.GetMetricName()) == 0 {
		hasError = true
		log.Errorln("name is required in metric config")
	}
	if len(m.GetMetricType()) == 0 {
		hasError = true
		log.Errorln("type is required in metric config")
	}
	switch m.GetMetricType() {
	case KeyMetricTypeHistogram:
		opts := m.GetOptions()
		var err error
		var iVal interface{}
		if iVal, err = opts.GetObject(OptKeyRextMetricDefHMetricBuckets); err != nil || iVal == nil {
			hasError = true
			log.Errorln("histogram metric should have some buckets defined")
		}
		buckets, okBuckets := iVal.([]float64)
		if !okBuckets {
			hasError = true
			log.WithFields(log.Fields{"key": OptKeyRextMetricDefHMetricBuckets, "val": iVal}).Errorln("error getting buckets values, histogram should have buckets defined")
		}
		if len(buckets) == 0 {
			hasError = true
			log.Errorln("histogram should have buckets defined")
		}
	case KeyMetricTypeCounter, KeyMetricTypeGauge:
	case KeyMetricTypeSummary:
		hasError = true
		log.Errorf("type %s is not supported yet\n", KeyMetricTypeSummary)
	default:
		hasError = true
		log.WithFields(log.Fields{"current": m.GetMetricType(), "expected": []string{KeyMetricTypeCounter, KeyMetricTypeGauge, KeyMetricTypeSummary, KeyMetricTypeHistogram}}).Errorln("invalid metric kind")
	}
	if m.GetNodeSolver() == nil {
		hasError = true
		log.Errorln("node solver is required in metric config")
	} else if m.GetNodeSolver().Validate() {
		hasError = true
	}
	for _, label := range m.GetLabels() {
		if label.Validate() {
			hasError = true
		}
	}
	return hasError
}

// ValidateRoot check if the root instance in parameter fill the required constraints
// to be considered as a valid RextRoot.
// Return true if any error is found
func ValidateRoot(r RextRoot) (hasError bool) {
	for _, srv := range r.GetServices() {
		if srv.Validate() {
			hasError = true
		}
	}
	return hasError
}
