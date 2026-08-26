package model

import "errors"

// 领域错误定义。所有错误均带可读消息，HTTP 层映射为状态码。
var (
	// ErrEmptyID 实体 ID 为空。
	ErrEmptyID = errors.New("id 不能为空")
	// ErrEmptyName 名称不能为空。
	ErrEmptyName = errors.New("名称不能为空")
	// ErrBadSegmentKind 非法区段类型。
	ErrBadSegmentKind = errors.New("非法区段类型")
	// ErrBadLength 区段长度非法。
	ErrBadLength = errors.New("区段长度必须为正数")
	// ErrBadSegmentState 非法区段状态。
	ErrBadSegmentState = errors.New("非法区段状态")
	// ErrBadSwitchPosition 非法道岔位置。
	ErrBadSwitchPosition = errors.New("非法道岔位置")
	// ErrBadSwitchEnds 道岔端点非法。
	ErrBadSwitchEnds = errors.New("道岔两端区段不能为空")
	// ErrSwitchSelfLoop 道岔两端指向同一区段。
	ErrSwitchSelfLoop = errors.New("道岔两端不能指向同一区段")
	// ErrUnknownSwitchPos 道岔位置未知，不能执行位置设置。
	ErrUnknownSwitchPos = errors.New("道岔位置未知，无法设置")
	// ErrBadRouteEnds 进路起终点非法。
	ErrBadRouteEnds = errors.New("进路起终点不能为空")
	// ErrBadPath 进路路径过短。
	ErrBadPath = errors.New("进路路径至少包含两个区段")
	// ErrBadPathEnds 路径首尾与起终点不一致。
	ErrBadPathEnds = errors.New("路径首尾与起终点不一致")
	// ErrSegmentSelfLoop 路径中区段重复。
	ErrSegmentSelfLoop = errors.New("进路路径不能包含重复区段")
	// ErrBadSwitchReq 道岔要求非法。
	ErrBadSwitchReq = errors.New("道岔要求非法")
	// ErrEmptyRelease 释放条件为空。
	ErrEmptyRelease = errors.New("释放条件至少需要一个依赖")
	// ErrBadVersionState 非法版本状态。
	ErrBadVersionState = errors.New("非法版本状态")
	// ErrBadTransition 非法状态迁移。
	ErrBadTransition = errors.New("非法状态迁移")
	// ErrEmptyVersion 版本 ID 为空。
	ErrEmptyVersion = errors.New("版本不能为空")
	// ErrBadConflictKind 非法冲突类型。
	ErrBadConflictKind = errors.New("非法冲突类型")
	// ErrBadConflictState 非法冲突状态。
	ErrBadConflictState = errors.New("非法冲突状态")
	// ErrEmptyRoute 进路 ID 为空。
	ErrEmptyRoute = errors.New("进路不能为空")
	// ErrEmptyConflict 冲突 ID 为空。
	ErrEmptyConflict = errors.New("冲突不能为空")
	// ErrBadExceptionState 非法例外状态。
	ErrBadExceptionState = errors.New("非法例外状态")
	// ErrEmptyReason 例外理由不能为空。
	ErrEmptyReason = errors.New("例外理由不能为空")
	// ErrBadSnapshotState 非法快照状态。
	ErrBadSnapshotState = errors.New("非法快照状态")

	// ErrNotClear 区段未空闲，无法锁闭。
	ErrNotClear = errors.New("区段未空闲，无法锁闭")
	// ErrNotReserved 区段未保留，无法释放保留。
	ErrNotReserved = errors.New("区段未处于保留状态")
	// ErrReservedRelease 区段处于保留中，不能直接出清。
	ErrReservedRelease = errors.New("区段处于保留中，不能直接出清")
	// ErrStateUnknown 区段状态未知。
	ErrStateUnknown = errors.New("区段状态未知")

	// ErrNotFound 实体不存在。
	ErrNotFound = errors.New("实体不存在")
	// ErrAlreadyExists 实体已存在。
	ErrAlreadyExists = errors.New("实体已存在")
	// ErrSealed 版本已封存，禁止修改。
	ErrSealed = errors.New("版本已封存，禁止修改")
	// ErrNotDraft 版本不是草稿态，禁止编辑。
	ErrNotDraft = errors.New("版本不是草稿态，禁止编辑")
	// ErrSwitchMissing 进路引用了不存在的道岔。
	ErrSwitchMissing = errors.New("进路引用了不存在的道岔")
	// ErrSegmentMissing 进路引用了不存在的区段。
	ErrSegmentMissing = errors.New("进路引用了不存在的区段")
	// ErrVersionNotValidated 版本尚未验证。
	ErrVersionNotValidated = errors.New("版本尚未完成验证")
	// ErrConflictResolved 冲突已解决，不能再裁决例外。
	ErrConflictResolved = errors.New("冲突已解决")
	// ErrPreconditionGap 释放条件引用了非路径区段（悬空依赖）。
	ErrPreconditionGap = errors.New("释放条件引用了非路径区段")
	// ErrSwitchContention 多条进路对同一道岔要求冲突。
	ErrSwitchContention = errors.New("多条进路对同一道岔要求冲突")
	// ErrOccupied 区段已被占用。
	ErrOccupied = errors.New("区段已被占用")
	// ErrUnknownSegment 区段不存在。
	ErrUnknownSegment = errors.New("区段不存在")
)
