package orchestrator

import (
	"go/ast"
	"go/parser"
	"go/token"
	"strconv"
	"strings"
)

type Tree struct {
	Root   *Node
	Flag   uint8 // 1 - processing, 2 - done, 3 - expression with division by zero
	Result float64
}

type Node struct {
	Val    interface{}
	Flag   uint8 // 1 - waiting (cant be done), 2 - can be done, 3 - args, 4 - processing, 5 - might be skipped
	Left   *Node
	Right  *Node
	Parent *Node
}

func operationChanger(op token.Token) uint8 {
	switch op {
	case token.ADD:
		return '+'
	case token.SUB:
		return '-'
	case token.MUL:
		return '*'
	case token.QUO:
		return '/'
	default:
		panic("invalid operation")
	}
}

func numChanger(num string) float64 {
	res, err := strconv.ParseFloat(num, 64)
	if err != nil {
		panic("invalid num when parse expr to tree")
	}
	return res
}

func step(cur **Node) {
	if (*cur).Left == nil {
		(*cur).Left = &Node{Parent: *cur}
		*cur = (*cur).Left
	} else {
		(*cur).Right = &Node{Parent: *cur}
		*cur = (*cur).Right
	}
}

func fillFlags(node *Node, list *[]*Node) {
	if node == nil {
		return
	}

	if node.Left == nil && node.Right == nil {
		node.Flag = 3
		return
	} else if node.Left == nil || node.Right == nil {
		panic("Non-full tree")
	}
	if node.Left.Flag == 0 {
		fillFlags(node.Left, list)
	}
	if node.Right.Flag == 0 {
		fillFlags(node.Right, list)
	}
	if node.Left.Flag == 3 && node.Right.Flag == 3 {
		node.Flag = 2
		*list = append(*list, node)
	} else {
		node.Flag = 1
	}
}

func NewTree(expr string) (*Tree, *[]*Node) {
	astTree, err := parser.ParseExpr(strings.ReplaceAll(expr, " ", ""))
	if err != nil {
		return &Tree{Flag: 15}, nil
	}

	var head *Node
	cur := head

	ast.Inspect(astTree, func(n ast.Node) bool {
		if head == nil {
			switch node := n.(type) {
			case *ast.BasicLit:
				head = &Node{Val: numChanger(node.Value)}
				cur = head
			case *ast.BinaryExpr:
				head = &Node{Val: operationChanger(node.Op)}
				cur = head
			case *ast.ParenExpr:
			default:
				head = &Node{Flag: 15}
			}
			return true
		}

		if head.Flag == 15 {
			return false
		}

		if n == nil {
			if cur == nil {
				return false
			}
			if cur.Flag == 0 {
				cur = cur.Parent
			} else {
				cur.Flag--
			}
			return true
		}

		switch node := n.(type) {
		case *ast.ParenExpr:
			cur.Flag++
		case *ast.BinaryExpr:
			step(&cur)
			cur.Val = operationChanger(node.Op)
		case *ast.BasicLit:
			step(&cur)
			cur.Val = numChanger(node.Value)
		default:
			head.Flag = 15
		}
		return true
	})

	if head.Flag == 15 {
		return &Tree{Flag: 15}, nil
	}

	list := make([]*Node, 0)
	fillFlags(head, &list)
	return &Tree{Root: head}, &list
}
