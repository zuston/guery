package Executor

import (
	"fmt"
	"io"

	"github.com/vmihailenco/msgpack"
	"github.com/xitongsys/guery/EPlan"
	"github.com/xitongsys/guery/Logger"
	"github.com/xitongsys/guery/Metadata"
	"github.com/xitongsys/guery/Row"
	"github.com/xitongsys/guery/Util"
	"github.com/xitongsys/guery/pb"
)

func (self *Executor) SetInstructionFilter(instruction *pb.Instruction) (err error) {
	var enode EPlan.EPlanFilterNode
	if err = msgpack.Unmarshal(instruction.EncodedEPlanNodeBytes, &enode); err != nil {
		return err
	}
	self.Instruction = instruction
	self.EPlanNode = &enode
	self.InputLocations = []*pb.Location{&enode.Input}
	self.OutputLocations = []*pb.Location{&enode.Output}
	return nil
}

func (self *Executor) RunFilter() (err error) {
	defer self.Clear()

	if self.Instruction == nil {
		return fmt.Errorf("No Instruction")
	}
	enode := self.EPlanNode.(*EPlan.EPlanFilterNode)

	md := &Metadata.Metadata{}
	reader := self.Readers[0]
	writer := self.Writers[0]
	if err = Util.ReadObject(reader, md); err != nil {
		return err
	}

	//write metadata
	if err = Util.WriteObject(writer, md); err != nil {
		return err
	}

	rbReader := Row.NewRowsBuffer(md, reader, nil)
	rbWriter := Row.NewRowsBuffer(md, nil, writer)

	//write rows
	var row *Row.Row
	var rg *Row.RowsGroup
	for {
		row, err = rbReader.ReadRow()
		if err == io.EOF {
			err = nil
			break
		}
		if err != nil {
			return err
		}
		rg = Row.NewRowsGroup(md)
		rg.Write(row)
		flag := true
		for _, booleanExpression := range enode.BooleanExpressions {
			rg.Reset()
			if ok, err := booleanExpression.Result(rg); !ok.(bool) && err == nil {
				flag = false
				break
			} else if err != nil {
				return err
			}
		}

		if flag {
			if err = rbWriter.WriteRow(row); err != nil {
				return err
			}
		}
	}

	if err = rbWriter.Flush(); err != nil {
		return err
	}

	Logger.Infof("RunFilter finished")
	return err
}
