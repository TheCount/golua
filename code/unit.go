package code

import (
	"fmt"
	"io"
)

type Unit struct {
	Code      []Opcode
	Constants []Constant
}

func NewUnit(code []Opcode, constants []Constant) *Unit {
	return &Unit{
		Code:      code,
		Constants: constants,
	}
}

type UnitDisassembler struct {
	unit   *Unit
	labels map[int]string
}

func NewUnitDisassembler(unit *Unit) *UnitDisassembler {
	return &UnitDisassembler{
		unit:   unit,
		labels: make(map[int]string),
	}
}

func (d *UnitDisassembler) SetLabel(offset int, lbl string) {
	d.labels[offset] = lbl
}

func (d *UnitDisassembler) GetLabel(offset int) string {
	lbl, ok := d.labels[offset]
	if !ok {
		lbl = fmt.Sprintf("L%d", len(d.labels))
		d.labels[offset] = lbl
	}
	return lbl
}

func (d *UnitDisassembler) ShortKString(ki uint16) string {
	k := d.unit.Constants[ki]
	return k.ShortString()
}

func (d *UnitDisassembler) Disassemble(w io.Writer) {
	disCode := make([]string, len(d.unit.Code))
	for i, opcode := range d.unit.Code {
		disCode[i] = opcode.Disassemble(d, i)
	}
	for i, dis := range disCode {
		fmt.Fprintf(w, "%s\t%d\t%08x\t%s\n", d.labels[i], i, d.unit.Code[i], dis)
	}
}
