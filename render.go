package gopkg

import (
	"encoding/json"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/luenci/errors"
)

// Result Request the result structure.
type Result struct {
	Data interface{} `json:"data"` // 返回内容
	Msg  string      `json:"msg"`  // 请求结果
	Code int         `json:"code"` // 状态码
}

func (r *Result) reset() {
	r.Data = nil
	r.Msg = ""
	r.Code = 0
}

var resultPool = &sync.Pool{
	New: func() interface{} {
		return new(Result)
	},
}

// Response response method.
func Response(ctx *gin.Context, code int, data interface{}) {
	// 状态码: 根据code（服务码）规则生成 eg: Code:200001, httpStatus: 200
	httpStatus := int(code / 1000)
	result := resultPool.Get().(*Result)
	defer func() {
		result.reset()
		resultPool.Put(result)
	}()

	var err error
	if _, ok := data.(error); ok {
		err = data.(error)
	}
	if err != nil {
		coder := errors.ParseCoder(err)
		result.Code = coder.Code()
		if coder.String() != "" {
			result.Msg = coder.String()
		} else {
			result.Msg = err.Error()
		}
	} else {
		result.Code = code
		result.Msg = "Success"
	}

	switch {
	case httpStatus >= 400 && httpStatus < 500:
		ctx.Abort()
		ctx.Set("warn", err)
		result.Data = err.Error()
	case httpStatus >= 500:
		ctx.Abort()
		ctx.Set("error", err)
	default:
		result.Data = data
	}
	b, _ := json.Marshal(&result)
	ctx.Set("ResponseBody", b)
	ctx.JSON(httpStatus, result)
}
