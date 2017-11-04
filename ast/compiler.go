package ast

import (
	"fmt"

	"github.com/arnodel/golua/ir"
)

type LexicalContext []map[Name]ir.Register

func (c LexicalContext) GetRegister(name Name) (reg ir.Register, ok bool) {
	for i := len(c) - 1; i >= 0; i-- {
		reg, ok = c[i][name]
		if ok {
			break
		}
	}
	return
}

func (c LexicalContext) AddToRoot(name Name, reg ir.Register) (ok bool) {
	ok = len(c) > 0
	if ok {
		c[0][name] = reg
	}
	return
}

func (c LexicalContext) AddToTop(name Name, reg ir.Register) (ok bool) {
	ok = len(c) > 0
	if ok {
		c[len(c)-1][name] = reg
	}
	return
}

func (c LexicalContext) PushNew() LexicalContext {
	return append(c, make(map[Name]ir.Register))
}

func (c LexicalContext) Pop() (LexicalContext, map[Name]ir.Register) {
	if len(c) == 0 {
		return c, nil
	}
	return c[:len(c)-1], c[len(c)-1]
}

func (c LexicalContext) Top() map[Name]ir.Register {
	if len(c) > 0 {
		return c[len(c)-1]
	}
	return nil
}

func (c LexicalContext) Dump() {
	for i, ns := range c {
		fmt.Printf("NS %d:\n", i)
		for name, reg := range ns {
			fmt.Printf("  %s: %s\n", name, reg)
		}
	}
}

type Compiler struct {
	registers []int
	context   LexicalContext
	parent    *Compiler
	upvalues  []ir.Register
	code      []ir.Instruction
	constants []ir.Constant
}

func NewCompiler(parent *Compiler) *Compiler {
	return &Compiler{
		context: LexicalContext{}.PushNew(),
		parent:  parent,
	}
}

func (c *Compiler) Dump() {
	fmt.Println("--context")
	c.context.Dump()
	fmt.Println("--constants")
	for i, k := range c.constants {
		fmt.Printf("k%d: %s\n", i, k)
	}
	fmt.Println("--code")
	for instr := range c.code {
		fmt.Println(instr)
	}
}

func (c *Compiler) GetRegister(name Name) (reg ir.Register, ok bool) {
	reg, ok = c.context.GetRegister(name)
	if ok || c.parent == nil {
		return
	}
	reg, ok = c.parent.GetRegister(name)
	if ok {
		c.upvalues = append(c.upvalues, reg)
		reg = ir.Register(-len(c.upvalues))
		c.context.AddToRoot(name, reg)
	}
	return
}

func (c *Compiler) GetFreeRegister() ir.Register {
	for i, n := range c.registers {
		if n == 0 {
			c.registers[i]++
			return ir.Register(i)
		}
	}
	c.registers = append(c.registers, 0)
	return ir.Register(len(c.registers) - 1)
}

func (c *Compiler) TakeRegister(reg ir.Register) {
	if int(reg) >= 0 {
		c.registers[reg]++
	}
}

func (c *Compiler) ReleaseRegister(reg ir.Register) {
	if c.registers[reg] == 0 {
		panic("Register cannot be released")
	}
	c.registers[reg]--
}

func (c *Compiler) PushContext() {
	c.context = c.context.PushNew()
}

func (c *Compiler) PopContext() {
	context, top := c.context.Pop()
	if top == nil {
		panic("Cannot pop empty context")
	}
	c.context = context
	for _, reg := range top {
		c.ReleaseRegister(reg)
	}
}

func (c *Compiler) DeclareLocal(name Name, reg ir.Register) {
	c.TakeRegister(reg)
	c.context.AddToTop(name, reg)
}

func (c *Compiler) Emit(instr ir.Instruction) {
	fmt.Printf("Emit %s\n", instr)
	c.code = append(c.code, instr)
}

func (c *Compiler) GetConstant(k ir.Constant) uint {
	for i, kk := range c.constants {
		if k == kk {
			return uint(i)
		}
	}
	c.constants = append(c.constants, k)
	return uint(len(c.constants) - 1)
}

func EmitConstant(c *Compiler, k ir.Constant, reg ir.Register) {
	c.Emit(ir.LoadConst{Dst: reg, Kidx: c.GetConstant(k)})
}
