package server

import (
	svrTypes "github.com/frangdelsolar/cms-builder/cms-builder-server/pkg/server/types"
)

const (
	CtxTraceId          svrTypes.ContextParamKey = "traceId"
	CtxRequestStartTime svrTypes.ContextParamKey = "requestStartTime"
	CtxRequestLogger    svrTypes.ContextParamKey = "requestLogger"
)
