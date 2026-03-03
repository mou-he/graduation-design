package code

// code响应状态码

type Code int64

const (
	CodeSuccess Code = 1000

	// ... 用户相关 2000 ...
	CodeInvalidParams    Code = 2001
	CodeUserExist        Code = 2002
	CodeUserNotExist     Code = 2003
	CodeInvalidPassword  Code = 2004
	CodeNotMatchPassword Code = 2005
	CodeInvalidToken     Code = 2006
	CodeNotLogin         Code = 2007
	CodeInvalidCaptcha   Code = 2008
	CodeRecordNotFound   Code = 2009
	CodeIllegalPassword  Code = 2010
	CodeEmailExist       Code = 2011
	CodeEmailNotExist    Code = 2012
	CodeNoAuth           Code = 2013

	CodeForbidden Code = 3001

	CodeServerBusy Code = 4001

	// ... AI 模型相关 5000 ...
	AIModelNotFind    Code = 5001
	AIModelCannotOpen Code = 5002
	AIModelFail       Code = 5003

	// ... 语音相关 6000 ...
	TTSFail Code = 6001

	// ... 视频生成相关 7000 (新增) ...
	VideoSubmitFail   Code = 7001 // 提交任务失败（如MQ挂了，Redis写不进去）
	VideoTaskNotFound Code = 7002 // 任务ID不存在或已过期
	VideoGenFail      Code = 7003 // 视频生成失败（智谱那边报错了）
	VideoSensitive    Code = 7004 // 内容敏感，被拦截
)

var msg = map[Code]string{
	CodeSuccess: "success",

	CodeInvalidParams:    "请求参数错误",
	CodeUserExist:        "用户名已存在",
	CodeUserNotExist:     "用户不存在",
	CodeInvalidPassword:  "用户名或密码错误",
	CodeNotMatchPassword: "两次密码不一致",
	CodeInvalidToken:     "无效的Token",
	CodeNotLogin:         "用户未登录",
	CodeInvalidCaptcha:   "验证码错误",
	CodeRecordNotFound:   "记录不存在",
	CodeIllegalPassword:  "密码不合法",
	CodeEmailExist:       "邮箱已存在",
	CodeEmailNotExist:    "邮箱不存在",
	CodeNoAuth:           "无权限操作",

	CodeForbidden: "权限不足",

	CodeServerBusy: "服务繁忙",

	AIModelNotFind:    "模型不存在",
	AIModelCannotOpen: "无法打开模型",
	AIModelFail:       "模型运行失败",
	TTSFail:           "语音服务失败",

	// 新增的视频相关提示信息
	VideoSubmitFail:   "视频任务提交失败，请稍后重试",
	VideoTaskNotFound: "任务不存在或已过期",
	VideoGenFail:      "视频生成失败，请检查提示词或重试",
	VideoSensitive:    "生成内容包含敏感信息，请修改提示词",
}

func (code Code) Code() int64 {
	return int64(code)
}

// Msg 获取响应消息
func (code Code) Msg() string {
	if m, ok := msg[code]; ok {
		return m
	}
	return msg[CodeServerBusy]
}
