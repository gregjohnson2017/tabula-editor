package util

import (
	"time"

	"github.com/go-gl/gl/v2.1/gl"
	"github.com/gregjohnson2017/tabula-editor/pkg/perf"
)

type GLQuery struct {
	queryID [2]uint32
}

func StartGLQuery() GLQuery {
	var glq GLQuery
	gl.GenQueries(2, &glq.queryID[0])
	gl.QueryCounter(glq.queryID[0], gl.TIMESTAMP)
	return glq
}

func (glq GLQuery) Stop(key string) {
	gl.QueryCounter(glq.queryID[1], gl.TIMESTAMP)
	var available int32
	for available == 0 {
		gl.GetQueryObjectiv(glq.queryID[1], gl.QUERY_RESULT_AVAILABLE, &available)
	}
	var startTime, stopTime uint64
	gl.GetQueryObjectui64v(glq.queryID[0], gl.QUERY_RESULT, &startTime)
	gl.GetQueryObjectui64v(glq.queryID[1], gl.QUERY_RESULT, &stopTime)
	perf.RecordAverageTime(key, time.Duration(int64(stopTime-startTime)).Nanoseconds())

	gl.DeleteQueries(2, &glq.queryID[0])
}
