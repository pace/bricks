// Copyright © 2020 by PACE Telematics GmbH. All rights reserved.
// Created at 2020/12/16 by Vincent Landgraf

package redact

import "context"

// ContextTransfer copies a request representation from one context to another.
func ContextTransfer(ctx, targetCtx context.Context) context.Context {
	if redactor := Ctx(ctx); redactor != nil {
		return context.WithValue(targetCtx, patternRedactorKey{}, redactor)
	}
	return targetCtx
}
