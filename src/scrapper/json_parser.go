package scrapper

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/oliveagle/jsonpath"
	"github.com/simelo/rextporter/src/config"
	"github.com/simelo/rextporter/src/util"
	log "github.com/sirupsen/logrus"
)

// JSONParser is a custom body parser that can desserialize json data
type JSONParser struct {
}

func (p JSONParser) decodeBody(body []byte) (val interface{}, err error) {
	generalScopeErr := "error decoding body as json"
	if err = json.Unmarshal(body, &val); err != nil {
		errCause := fmt.Sprintf("can not decode the body: %s. Err: %s", string(body), err.Error())
		return nil, util.ErrorFromThisScope(errCause, generalScopeErr)
	}
	return val, err
}

func (p JSONParser) pathLookup(path string, val interface{}) (node interface{}, err error) {
	if len(path) == 0 {
		log.Errorln("node path is required")
		return nil, config.ErrKeyEmptyValue
	}
	generalScopeErr := "error looking for node in val"
	jPath := "$" + strings.Replace(path, "/", ".", -1)
	if node, err = jsonpath.JsonPathLookup(val, jPath); err != nil {
		errCause := fmt.Sprintln("can not locate the path: ", err.Error())
		return nil, util.ErrorFromThisScope(errCause, generalScopeErr)
	}
	return node, err
}
