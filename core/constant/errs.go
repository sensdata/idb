package constant

import "errors"

const (
	CodeFailed          = 0
	CodeSuccess         = 200
	CodeErrBadRequest   = 400
	CodeErrUnauthorized = 401
	CodeErrUnSafety     = 402
	CodeErrForbidden    = 403
	CodeErrNotFound     = 404
	CodePasswordExpired = 405
	CodeAuth            = 406
	CodeGlobalLoading   = 407

	CodeErrEnvironment = 410
	CodeErrIP          = 411
	CodeErrDomain      = 412
	CodeErrHeader      = 413
	CodeErrXpack       = 414

	CodeErrInternalServer = 500
)

// api
var (
	ErrInternalServer           = errors.New("InternalServerError")
	ErrInvalidParams            = errors.New("InvalidParams")
	ErrNotLogin                 = errors.New("NotLogin")
	ErrPasswordExpired          = errors.New("PasswordExpired")
	ErrInvalidOldPassword       = errors.New("InvalidOldPassword")
	ErrNameIsExist              = errors.New("NameIsExist")
	ErrInvalidAccountOrPassword = errors.New("InvalidAccountOrPassword")
	ErrAuth                     = errors.New("ErrAuth")
	ErrNoRecords                = errors.New("NoRecords")
	ErrRecordExist              = errors.New("ErrRecordExist")
	ErrRecordNotFound           = errors.New("ErrRecordNotFound")
	ErrStructTransform          = errors.New("ErrStructTransform")
	ErrBussiness                = errors.New("BussinessFailed")
	ErrHostNotFound             = errors.New("ErrHostNotFound")
	ErrHost                     = errors.New("HostError")
	ErrSsh                      = errors.New("SshError")
	ErrAgent                    = errors.New("AgentError")
	ErrFileNotExist             = errors.New("file does not exist")
)

// cmd
var (
	ErrCmdIllegal  = "ErrCmdIllegal"
	ErrCmdTimeout  = "ErrCmdTimeout"
	ErrCmdNotFound = "ErrCmdNotFound"
)

// common
var (
	ErrInUsed       = "ErrInUsed"
	ErrObjectInUsed = "ErrObjectInUsed"
	ErrPortInUsed   = "ErrPortInUsed"
	ErrPortRules    = "ErrPortRules"
	ErrNotInstalled = "ErrNotInstalled"
)

// file
var (
	ErrPathNotFound     = "ErrPathNotFound"
	ErrMovePathFailed   = "ErrMovePathFailed"
	ErrLinkPathNotFound = "ErrLinkPathNotFound"
	ErrFileIsExit       = "ErrFileIsExit"
	ErrFileNotFound     = "ErrFileNotFound"
	ErrFileUpload       = "ErrFileUpload"
	ErrFileDownloadDir  = "ErrFileDownloadDir"
	ErrFileOpen         = "ErrFileOpen"
	ErrFileRead         = "ErrFileRead"
	ErrFavoriteExist    = "ErrFavoriteExist"
	ErrFileCanNotRead   = "ErrFileCanNotRead"
	ErrFileToLarge      = "ErrFileToLarge"
)

// json
var (
	ErrJSONMarshal = errors.New("ErrJSONMarshal")
)

// docker
var (
	ErrContainerName = "ErrContainerName"
)
var (
	OperationFailed  = "Failed"
	OperationSuccess = "Success"
)

// session
var (
	ErrSessionLimit = "ErrSessionLimit"
)
