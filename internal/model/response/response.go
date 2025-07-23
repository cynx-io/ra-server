package response

import (
	"github.com/cynxees/cynx-core/src/response"
)

func setResponse[Resp response.Generic](resp Resp, code Code) {
	resp.GetBase().Code = code.String()
	resp.GetBase().Desc = responseCodeNames[code]
}

type OnError[Resp response.Generic] func(resp Resp)

func Success[Resp response.Generic](resp Resp) {
	setResponse(resp, codeSuccess)
}

func ErrorValidation[Resp response.Generic](resp Resp) {
	setResponse(resp, codeValidationError)
}

func ErrorUnauthorized[Resp response.Generic](resp Resp) {
	setResponse(resp, codeUnauthorized)
}

func ErrorNotAllowed[Resp response.Generic](resp Resp) {
	setResponse(resp, codeNotAllowed)
}

func ErrorNotFound[Resp response.Generic](resp Resp) {
	setResponse(resp, codeNotFound)
}

func ErrorInvalidCredentials[Resp response.Generic](resp Resp) {
	setResponse(resp, codeInvalidCredentials)
}

func ErrorInternal[Resp response.Generic](resp Resp) {
	setResponse(resp, codeInternalError)
}

func ErrorCanceled[Resp response.Generic](resp Resp) {
	setResponse(resp, codeCanceledError)
}

func ErrorDbTopic[Resp response.Generic](resp Resp) {
	setResponse(resp, codeDbTopicError)
}

func ErrorDbMode[Resp response.Generic](resp Resp) {
	setResponse(resp, codeDbModeError)
}

func ErrorDbAnswer[Resp response.Generic](resp Resp) {
	setResponse(resp, codeDbAnswer)
}

func ErrorDbAnswerCategory[Resp response.Generic](resp Resp) {
	setResponse(resp, codeDbAnswerCategory)
}

func ErrorDbDailyGame[Resp response.Generic](resp Resp) {
	setResponse(resp, codeDbDailyGame)
}

func ErrorDbDailyGameGuess[Resp response.Generic](resp Resp) {
	setResponse(resp, codeDbDailyGameGuess)
}

func ErrorAlreadyExists[Resp response.Generic](resp Resp) {
	setResponse(resp, codeAlreadyExists)
}
