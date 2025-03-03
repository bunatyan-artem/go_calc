package orchestrator

import (
	"go/ast"
	"go/parser"
	"strings"
)

type Tree struct {
	Root   *Node
	Flag   uint8
	Result float64
}

type Node struct {
	Val    interface{}
	Flag   uint8
	Left   *Node
	Right  *Node
	Parent *Node
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
	} else if node.Left == nil || node.Right == nil {
		panic("Non-full tree")
	} else if node.Left.Flag == 3 && node.Right.Flag == 3 {
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
		if head.Flag == 15 {
			return false
		}

		if head == nil {
			switch node := n.(type) {
			case *ast.BasicLit:
				head = &Node{Val: node.Value}
			case *ast.BinaryExpr:
				head = &Node{Val: node.Op}
			case *ast.ParenExpr:
			default:
				head.Flag = 15
			}
			return true
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
			cur.Val = node.Op
		case *ast.BasicLit:
			step(&cur)
			cur.Val = node.Value
		default:
			head.Flag = 15
		}
		return true
	})

	if head.Flag == 15 {
		return &Tree{Flag: 15}, nil
	}

	var list *[]*Node
	fillFlags(head, list)
	return &Tree{Root: head}, list
}
