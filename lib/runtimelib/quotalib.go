package runtimelib

import (
	"github.com/arnodel/golua/lib/packagelib"
	rt "github.com/arnodel/golua/runtime"
)

var LibLoader = packagelib.Loader{
	Load: load,
	Name: "runtime",
}

func load(r *rt.Runtime) rt.Value {
	pkg := rt.NewTable()
	r.SetEnvGoFunc(pkg, "callcontext", callcontext, 2, true)
	r.SetEnvGoFunc(pkg, "context", context, 0, false)

	createContextMetatable(r)

	return rt.TableValue(pkg)
}

func context(t *rt.Thread, c *rt.GoCont) (rt.Cont, *rt.Error) {
	ctx := newContextValue(t.Runtime, t.RuntimeContext())
	return c.PushingNext1(t.Runtime, ctx), nil
}

func callcontext(t *rt.Thread, c *rt.GoCont) (next rt.Cont, retErr *rt.Error) {
	quotas, err := c.TableArg(0)
	if err != nil {
		return nil, err.AddContext(c)
	}
	var (
		memQuotaV = quotas.Get(rt.StringValue("memlimit"))
		cpuQuotaV = quotas.Get(rt.StringValue("cpulimit"))
		ioV       = quotas.Get(rt.StringValue("io"))
		memQuota  int64
		cpuQuota  int64
		ok        bool
		f         = c.Arg(1)
		fArgs     = c.Etc()
		flags     rt.RuntimeContextFlags
	)
	if !rt.IsNil(memQuotaV) {
		memQuota, ok = memQuotaV.TryInt()
		if !ok {
			return nil, rt.NewErrorS("memquota must be an integer").AddContext(c)
		}
		if memQuota <= 0 {
			return nil, rt.NewErrorS("memquota must be positive").AddContext(c)
		}
	}
	if !rt.IsNil(cpuQuotaV) {
		cpuQuota, ok = cpuQuotaV.TryInt()
		if !ok {
			return nil, rt.NewErrorS("cpuquota must be an integer").AddContext(c)
		}
		if cpuQuota <= 0 {
			return nil, rt.NewErrorS("cpuquota must be positive").AddContext(c)
		}
	}
	if !rt.IsNil(ioV) {
		status, _ := ioV.TryString()
		switch status {
		case "off":
			flags |= rt.RCF_NoIO
		case "on":
			// Nothing to do
		default:
			return nil, rt.NewErrorS("io must be 'on' or 'off'").AddContext(c)
		}
	}

	// Push new quotas
	t.PushQuota(uint64(cpuQuota), uint64(memQuota), flags)

	next = c.Next()
	res := rt.NewTerminationWith(0, true)
	defer func() {

		// var (
		// 	memUsed, memQuota = t.MemQuotaStatus()
		// 	cpuUsed, cpuQuota = t.CPUQuotaStatus()
		// )
		// if memQuota > 0 {
		// 	t.SetEnv(quotas, "memused", rt.IntValue(int64(memUsed)))
		// 	t.SetEnv(quotas, "memquota", rt.IntValue(int64(memQuota)))
		// }
		// if cpuQuota > 0 {
		// 	t.SetEnv(quotas, "cpuused", rt.IntValue(int64(cpuUsed)))
		// 	t.SetEnv(quotas, "cpuquota", rt.IntValue(int64(cpuQuota)))
		// }
		// // In any case, pop the quotas
		// t.PopQuota()

		ctx := t.PopContext()
		if retErr != nil {
			// In this case there was an error, so no panic.  We return the
			// error normally. To avoid this a user can wrap f in a pcall.
			return
		}
		t.Push1(next, newContextValue(t.Runtime, ctx))
		if r := recover(); r != nil {
			_, ok := r.(rt.QuotaExceededError)
			if !ok {
				panic(r)
			}
		} else {
			t.Push(next, res.Etc()...)
		}
	}()
	retErr = rt.Call(t, f, fArgs, res)
	if retErr != nil {
		return nil, retErr
	}
	return
}
