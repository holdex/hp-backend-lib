package libhttp

import (
	"encoding/json"
	"fmt"
	"net/http"

	"bitbucket.org/holdex/hp-backend-lib/err"
)

const (
	errNotAuthorized    = "not_authorized"
	errNotAuthenticated = "not_authenticated"
	errMalformedBody    = "malformed_body"
	errUnknown          = "unknown"
	errInvalidArgument  = "invalid_argument"
)

// +--------------------------------------------------------------------------------------------------------------------
// | Requests
// +--------------------------------------------------------------------------------------------------------------------

func DecodeJSONBody(w http.ResponseWriter, r *http.Request, i interface{}) bool {
	decoder := json.NewDecoder(r.Body)
	defer r.Body.Close()
	if err := decoder.Decode(i); err != nil {
		Err400(w, errMalformedBody, err)
		return false
	}
	return true
}

// +--------------------------------------------------------------------------------------------------------------------
// | Responses
// +--------------------------------------------------------------------------------------------------------------------

func writeResponse(w http.ResponseWriter, code int, i interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	jr, _ := json.Marshal(i)
	w.Write(jr)
}

func errorResponse(reason string, details interface{}) map[string]interface{} {
	r := map[string]interface{}{
		"status": "error",
		"reason": reason,
	}

	if details != nil {
		r["details"] = fmt.Sprint(details)
	}

	return r
}

func Ok(w http.ResponseWriter) {
	writeResponse(w, http.StatusOK, map[string]interface{}{
		"status": "success",
	})
}

func OkData(w http.ResponseWriter, data interface{}) {
	writeResponse(w, http.StatusOK, data)
}

func ErrUnknown(w http.ResponseWriter, err error) {
	switch e := err.(type) {
	case *liberr.InvalidArgument:
		Err400(w, errInvalidArgument, e)
	case *liberr.NotAuthorized:
		Err401(w, errNotAuthorized, e)
	case *liberr.NotAuthenticated:
		OkErr(w, errNotAuthenticated, e)
	default:
		Err500(w, e)
	}
}

func OkErr(w http.ResponseWriter, reason string, details interface{}) {
	writeResponse(w, http.StatusOK, errorResponse(reason, details))
}

func Err400(w http.ResponseWriter, reason string, details interface{}) {
	writeResponse(w, http.StatusBadRequest, errorResponse(reason, details))
}

func Err401(w http.ResponseWriter, reason string, details interface{}) {
	writeResponse(w, http.StatusUnauthorized, errorResponse(reason, details))
}

func Err500(w http.ResponseWriter, details ...interface{}) {
	writeResponse(w, http.StatusInternalServerError, errorResponse(errUnknown, fmt.Sprintln(details...)))
}

func setCookie(w http.ResponseWriter, name, value string, expires int, httpOnly bool) {
	http.SetCookie(w, &http.Cookie{
		Name:     name,
		Value:    value,
		Path:     "/",
		MaxAge:   expires,
		HttpOnly: httpOnly,
	})
}
