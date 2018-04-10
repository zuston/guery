package Plan

import (
	"fmt"

	"github.com/antlr/antlr4/runtime/Go/antlr"
	"github.com/xitongsys/guery/Common"
	"github.com/xitongsys/guery/Context"
	"github.com/xitongsys/guery/DataSource"
	"github.com/xitongsys/guery/parser"
)

type PlanNodeType int32

const (
	_ PlanNodeType = iota
	SCANNODE
	JOINNODE
	FILTERNODE
	ORDERBYNODE
	LIMITNODE
	SELECTNODE
	UNIONNODE
	HAVINGNODE
	UNITNODE
)

type PlanNode interface {
	Execute() DataSource.DataSource
}

//////////////////
type PlanOrderByNode struct {
	Input PlanNode
}

func NewPlanOrderByNode(ctx *Context.Context, input PlanNode, sortItems []parser.ISortItemContext) *PlanOrderByNode {
	res := &PlanOrderByNode{
		Input: input,
	}
	return res
}

func (self *PlanOrderByNode) Execute() DataSource.DataSource {
	return self.Input.Execute()
}

/////////////////
type PlanLimitNode struct {
	Input       PlanNode
	LimitNumber *int64
}

func NewPlanLimitNode(ctx *Context.Context, input PlanNode, t antlr.TerminalNode) *PlanLimitNode {
	res := &PlanLimitNode{
		Input: input,
	}
	if ns := t.GetText(); ns != "ALL" {
		var num int64
		fmt.Sscanf(ns, "%d", &num)
		res.LimitNumber = &num
	}
	return res
}

func (self *PlanLimitNode) Execute() DataSource.DataSource {
	return self.Input.Execute()
}

////////////////
type PlanUnionNode struct {
	LeftInput  PlanNode
	RightInput PlanNode
	Operator   Common.Operator
}

func NewPlanUnionNode(ctx *Context.Context, left PlanNode, right PlanNode, op antlr.Token) *PlanUnionNode {
	var operator Common.Operator
	switch op.GetText() {
	case "INTERSECT":
		operator = Common.INTERSECT
	case "UNION":
		operator = Common.UNION
	case "EXCEPT":
		operator = Common.EXCEPT
	}

	res := &PlanUnionNode{
		LeftInput:  left,
		RightInput: right,
		Operator:   operator,
	}
	return res
}

func (self *PlanUnionNode) Execute() DataSource.DataSource {
	return self.LeftInput.Execute()
}

//////////////
type PlanFiliterNode struct {
	Input PlanNode
}

func NewPlanFiliterNode(ctx *Context.Context, input PlanNode, t parser.IBooleanExpressionContext) *PlanFiliterNode {
	res := &PlanFiliterNode{
		Input: input,
	}
	return res
}

func (self *PlanFiliterNode) Execute() DataSource.DataSource {
	return self.Input.Execute()
}

//////////////
type PlanHavingNode struct {
	Input PlanNode
}

func NewPlanHavingNode(ctx *Context.Context, input PlanNode, t parser.IBooleanExpressionContext) *PlanHavingNode {
	res := &PlanHavingNode{
		Input: input,
	}
	return res
}

func (self *PlanHavingNode) Execute() DataSource.DataSource {
	return self.Input.Execute()
}

////////////////
type PlanSelectNode struct {
	Input       PlanNode
	SelectItems []*SelectItemNode
}

func NewPlanSelectNode(ctx *Context.Context, input PlanNode, items []parser.ISelectItemContext, groupBy parser.IGroupByContext) *PlanSelectNode {

	res := &PlanSelectNode{
		Input:       input,
		SelectItems: []*SelectItemNode{},
	}
	for i := 0; i < len(items); i++ {
		res.SelectItems = append(res.SelectItems, NewSelectItemNode(ctx, items[i]))
	}
	return res
}

func (self *PlanSelectNode) Execute() DataSource.DataSource {
	var ds DataSource.DataSource
	if self.Input != nil {
		ds = self.Input.Execute()
	}

	names := []string{}
	size := ds.Size()
	for i := 0; i < len(self.SelectItems); i++ {
		item := self.SelectItems[i]
		names = append(names, item.GetNames()...)
	}
	tb := DataSource.NewTableSource("", names)
	for i := int64(0); i < size; i++ {
		vals := []interface{}{}
		for j := 0; j < len(self.SelectItems); j++ {
			item := self.SelectItems[j]
			vals = append(vals, item.Result(ds)...)
		}
		tb.Append(vals)
	}

	return tb
}

///////////////////
type PlanScanNode struct {
	Input DataSource.DataSource
}

func NewPlanScanNodeFromDataSource(ctx *Context.Context, input DataSource.DataSource) *PlanScanNode {
	res := &PlanScanNode{
		Input: input,
	}
	return res
}

func NewPlanScanNode(ctx *Context.Context, name string) *PlanScanNode {
	res := &PlanScanNode{
		Input: ctx.Tables[name],
	}
	return res
}

func (self *PlanScanNode) Execute() DataSource.DataSource {
	return self.Input
}
