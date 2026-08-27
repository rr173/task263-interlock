package httpapi

import (
	"errors"
	"net/http"

	"task263-interlock/internal/model"
)

// mapError 将领域错误映射为 HTTP 状态码。
func mapError(err error) (int, string) {
	switch {
	case err == nil:
		return http.StatusOK, ""
	case errors.Is(err, model.ErrNotFound):
		return http.StatusNotFound, err.Error()
	case errors.Is(err, model.ErrAlreadyExists):
		return http.StatusConflict, err.Error()
	case errors.Is(err, model.ErrBadTransition),
		errors.Is(err, model.ErrSealed),
		errors.Is(err, model.ErrNotDraft),
		errors.Is(err, model.ErrVersionNotValidated),
		errors.Is(err, model.ErrConflictResolved),
		errors.Is(err, model.ErrNotClear),
		errors.Is(err, model.ErrNotReserved),
		errors.Is(err, model.ErrReservedRelease),
		errors.Is(err, model.ErrStateUnknown),
		errors.Is(err, model.ErrPreconditionGap),
		errors.Is(err, model.ErrSwitchContention),
		errors.Is(err, model.ErrOccupied):
		return http.StatusConflict, err.Error()
	case errors.Is(err, model.ErrEmptyID),
		errors.Is(err, model.ErrEmptyName),
		errors.Is(err, model.ErrBadSegmentKind),
		errors.Is(err, model.ErrBadLength),
		errors.Is(err, model.ErrBadSegmentState),
		errors.Is(err, model.ErrBadSwitchPosition),
		errors.Is(err, model.ErrBadSwitchEnds),
		errors.Is(err, model.ErrSwitchSelfLoop),
		errors.Is(err, model.ErrUnknownSwitchPos),
		errors.Is(err, model.ErrBadRouteEnds),
		errors.Is(err, model.ErrBadPath),
		errors.Is(err, model.ErrBadPathEnds),
		errors.Is(err, model.ErrSegmentSelfLoop),
		errors.Is(err, model.ErrBadSwitchReq),
		errors.Is(err, model.ErrEmptyRelease),
		errors.Is(err, model.ErrBadVersionState),
		errors.Is(err, model.ErrEmptyVersion),
		errors.Is(err, model.ErrBadConflictKind),
		errors.Is(err, model.ErrBadConflictState),
		errors.Is(err, model.ErrEmptyRoute),
		errors.Is(err, model.ErrEmptyConflict),
		errors.Is(err, model.ErrBadExceptionState),
		errors.Is(err, model.ErrEmptyReason),
		errors.Is(err, model.ErrBadSnapshotState),
		errors.Is(err, model.ErrSwitchMissing),
		errors.Is(err, model.ErrSegmentMissing):
		return http.StatusBadRequest, err.Error()
	default:
		return http.StatusInternalServerError, "内部错误: " + err.Error()
	}
}

// respond 统一错误响应。
func respond(w http.ResponseWriter, err error) {
	code, msg := mapError(err)
	if code == http.StatusOK {
		return
	}
	writeErr(w, code, msg)
}
