package ast

import "github.com/arnodel/golua/ir"

type FunctionCall struct {
	target ExpNode
	method Name
	args   []ExpNode
}

func (f FunctionCall) HWrite(w HWriter) {
	w.Writef("call")
	w.Indent()
	w.Next()
	w.Writef("target: ")
	// w.Indent()
	f.target.HWrite(w)
	// w.Dedent()
	if f.method != "" {
		w.Next()
		w.Writef("method: %s", f.method)
	}
	for i, arg := range f.args {
		w.Next()
		w.Writef("arg_%d: ", i)
		arg.HWrite(w)
	}
	w.Dedent()
}

func (f FunctionCall) CompileExp(c *Compiler, dst ir.Register) ir.Register {
	f.CompileCall(c)
	c.Emit(ir.Receive{Dst: []ir.Register{dst}})
	return dst
}

func CallWithArgs(c *Compiler, args []ExpNode, fReg ir.Register) {
	c.TakeRegister(fReg)
	for i, arg := range args {
		var argReg ir.Register
		argReg = c.GetFreeRegister()
		if i == len(args)-1 {
			argFc, ok := arg.(FunctionCall)
			if ok {
				argFc.CompileCall(c)
				argReg = c.GetFreeRegister()
				c.Emit(ir.ReceiveEtc{Etc: argReg})
				goto PushLbl
			}
		}
		argReg = CompileExp(c, arg)
	PushLbl:
		c.Emit(ir.Push{Cont: fReg, Item: argReg})
	}
	c.Emit(ir.Call{Cont: fReg})
	c.ReleaseRegister(fReg)
}

func (f FunctionCall) CompileCall(c *Compiler) {
	fReg := CompileExp(c, f.target)
	if f.method != "" {
		self := fReg
		c.TakeRegister(self)
		fReg = c.GetFreeRegister()
		mReg := c.GetFreeRegister()
		EmitConstant(c, ir.String(f.method), mReg)
		c.Emit(ir.Lookup{
			Dst:   fReg,
			Table: self,
			Index: mReg,
		})
		c.Emit(ir.MkCont{Dst: fReg, Closure: fReg})
		c.Emit(ir.Push{Cont: fReg, Item: self})
		c.ReleaseRegister(self)
	} else {
		c.Emit(ir.MkCont{Dst: fReg, Closure: fReg})
	}
	c.Emit(ir.PushCC{Cont: fReg})
	CallWithArgs(c, f.args, fReg)
}

func (f FunctionCall) CompileStat(c *Compiler) {
	f.CompileCall(c)
	c.Emit(ir.Receive{})
}

func NewFunctionCall(target ExpNode, method Name, args []ExpNode) (*FunctionCall, error) {
	return &FunctionCall{
		target: target,
		method: method,
		args:   args,
	}, nil
}
