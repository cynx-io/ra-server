package response

type Code string

func (r Code) String() string {
	return string(r)
}

const (
	// Expected Error
	codeSuccess            Code = "00"
	codeValidationError    Code = "VE"
	codeUnauthorized       Code = "UA"
	codeNotAllowed         Code = "NA"
	codeNotFound           Code = "NF"
	codeInvalidCredentials Code = "IC"
	codeAlreadyExists      Code = "AE"

	// Internal
	codeInternalError Code = "I-IE"
	codeCanceledError Code = "I-CE"

	// External Errors
	// Database Errors
	codeDbTopicError     Code = "DB-TPC"
	codeDbModeError      Code = "DB-MDE"
	codeDbAnswer         Code = "DB-ANS"
	codeDbAnswerCategory Code = "DB-ANC"
	codeDbDailyGame      Code = "DB-DLY"
	codeDbDailyGameGuess Code = "DB-DLG"
)

var responseCodeNames = map[Code]string{
	// Expected Error
	codeSuccess:            "Success",
	codeValidationError:    "Validation Error",
	codeUnauthorized:       "Not Authorized",
	codeNotAllowed:         "Not Allowed",
	codeNotFound:           "Not Found",
	codeInvalidCredentials: "Invalid Credentials",
	codeAlreadyExists:      "Already Exists",

	// Internal
	codeInternalError: "Internal Error",
	codeCanceledError: "Canceled Error",

	// Database Errors
	codeDbTopicError:     "Database Topic Error",
	codeDbModeError:      "Database Mode Error",
	codeDbAnswer:         "Database Answer Error",
	codeDbAnswerCategory: "Database Answer Category Error",
	codeDbDailyGame:      "Database Daily Game Error",
	codeDbDailyGameGuess: "Database Daily Game Guess Error",
}
