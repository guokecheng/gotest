package response

import (
	"easy-echo/common"
	"github.com/labstack/echo"
	"net/http"
)

//HTTPResp ...
type HTTPResp struct {
	ErrCode int         `json:"errcode"`
	ErrMsg  string      `json:"errmsg"`
	Data    interface{} `json:"data"`
}

func ShowError(c echo.Context, apiError *common.APIError) error {
	return c.JSON(http.StatusOK, &HTTPResp{
		ErrCode: apiError.Code,
		ErrMsg:  apiError.Message,
		Data:    apiError.Data,
	})
}

func ShowSuccess(c echo.Context, msg string, data interface{}) error {
	return c.JSON(http.StatusOK, &HTTPResp{
		ErrCode: 0,
		ErrMsg:  msg,
		Data:    data,
	})
}
