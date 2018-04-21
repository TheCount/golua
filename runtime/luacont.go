package runtime

import (
	"errors"

	"github.com/arnodel/golua/code"
)

type LuaContinuation struct {
	*Closure
	registers []Value
	pc        int
}

func NewLuaContinuation(clos *Closure) *LuaContinuation {
	if clos.upvalueIndex < len(clos.upvalues) {
		panic("Closure not ready")
	}
	return &LuaContinuation{
		Closure:   clos,
		registers: make([]Value, clos.RegCount),
	}
}

func (c *LuaContinuation) Push(val Value) {
	opcode := c.code[c.pc]
	if opcode.HasType0() {
		c.pc++
		dst := opcode.GetA()
		if opcode.GetF() {
			// It's an etc
			// TODO
		}
		c.setReg(dst, val)
	}

}

func (c *LuaContinuation) setReg(reg code.Reg, val Value) {
	if val == nil {
		val = NilType{}
	}
	switch reg.Tp() {
	case code.Register:
		c.registers[reg.Idx()] = val
	default:
		c.upvalues[reg.Idx()] = val
	}
}

func (c *LuaContinuation) getReg(reg code.Reg) Value {
	switch reg.Tp() {
	case code.Register:
		return c.registers[reg.Idx()]
	default:
		return c.upvalues[reg.Idx()]
	}
}

func (c *LuaContinuation) RunInThread(t *Thread) (Continuation, error) {
	pc := c.pc
	consts := c.consts
RunLoop:
	for pc < len(c.code) {
		opcode := c.code[pc]
		if opcode.HasType1() {
			dst := opcode.GetA()
			x := c.getReg(opcode.GetB())
			y := c.getReg(opcode.GetC())
			var res Value
			var err error
			switch opcode.GetX() {
			case code.OpAdd:
				res, err = add(t, x, y)
			case code.OpSub:
				res, err = sub(t, x, y)
			case code.OpMul:
				res, err = mul(t, x, y)
			case code.OpDiv:
				res, err = div(t, x, y)
			case code.OpFloorDiv:
				res, err = idiv(t, x, y)
			case code.OpMod:
				res, err = mod(t, x, y)
			case code.OpPow:
				res, err = pow(t, x, y)
			case code.OpBitAnd:
				panic("unimplemented")
			case code.OpBitOr:
				panic("unimplemented")
			case code.OpBitXor:
				panic("unimplemented")
			case code.OpShiftL:
				panic("unimplemented")
			case code.OpShiftR:
				panic("unimplemented")
			case code.OpEq:
				panic("unimplemented")
			case code.OpLt:
				panic("unimplemented")
			case code.OpLeq:
				panic("unimplemented")
			case code.OpConcat:
				panic("unimplemented")
			default:
				panic("unsupported")
			}
			if err != nil {
				return nil, err
			}
			c.setReg(dst, res)
			pc++
			continue RunLoop
		}
		switch opcode.TypePfx() {
		case code.Type0Pfx:
			dst := opcode.GetA()
			if opcode.GetF() {
				// It's an etc, do something else instead?
			}
			c.setReg(dst, nil)
		case code.Type2Pfx:
			reg := opcode.GetA()
			coll := c.getReg(opcode.GetB())
			idx := c.getReg(opcode.GetC())
			if !opcode.GetF() {
				val, err := getindex(t, coll, idx)
				if err != nil {
					return nil, err
				}
				c.setReg(reg, val)
			} else {
				err := setindex(t, coll, idx, c.getReg(reg))
				if err != nil {
					return nil, err
				}
			}
			pc++
			continue RunLoop
		case code.Type3Pfx:
			n := opcode.GetN()
			var val Value
			switch code.UnOpK16(opcode.GetY()) {
			case code.OpK:
				val = consts[n]
			case code.OpClosureK:
				val = NewClosure(consts[n].(*Code))
			default:
				panic("Unsupported opcode")
			}
			dst := opcode.GetA()
			if opcode.GetF() {
				// dst must contain a continuation
				cont := c.getReg(dst).(Continuation)
				cont.Push(val)
			} else {
				c.setReg(dst, val)
			}
			pc++
			continue RunLoop
		case code.Type4Pfx:
			dst := opcode.GetA()
			var res Value
			var err error
			if opcode.HasType4a() {
				val := c.getReg(opcode.GetB())
				switch opcode.GetZ() {
				case code.OpNeg:
					res, err = unm(t, val)
				case code.OpBitNot:
					panic("unimplemented")
				case code.OpLen:
					panic("unimplemented")
				case code.OpClosure:
					panic("unimplemented")
				case code.OpCont:
					panic("unimplemented")
				case code.OpId:
					panic("unimplemented")
				case code.OpTruth:
					panic("unimplemented")
				case code.OpCell:
					panic("unimplemented")
				case code.OpNot:
					panic("unimplemented")
				case code.OpUpvalue:
					c.getReg(dst).(*Closure).AddUpvalue(val)
					pc++
					continue RunLoop
				default:
					panic("unsupported")
				}
			} else {
				// Type 4b
				switch code.UnOpK(opcode.GetZ()) {
				case code.OpCC:
					res = c
				case code.OpTable:
					res = NewTable()
				default:
					panic("unsupported")
				}
			}
			if err != nil {
				return nil, err
			}
			if opcode.GetF() {
				c.getReg(dst).(Continuation).Push(res)
			} else {
				c.setReg(dst, res)
			}
			pc++
			continue RunLoop
		case code.Type5Pfx:
			switch code.JumpOp(opcode.GetY()) {
			case code.OpJump:
				pc += int(opcode.GetN())
				continue RunLoop
			case code.OpJumpIf:
				test := truth(c.getReg(opcode.GetA()))
				if test == opcode.GetF() {
					pc += int(opcode.GetN())
				}
				continue RunLoop
			case code.OpCall:
				pc++
				return c.getReg(opcode.GetA()).(Continuation), nil
			}
		}
	}
	return nil, errors.New("Invalid PC")
}
