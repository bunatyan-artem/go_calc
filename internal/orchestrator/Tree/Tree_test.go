package Tree

import (
	"testing"
)

func TestNewTree(t *testing.T) {
	tests := []struct {
		expr string
	}{
		{"2+2"},
		{"5-3"},
		{" 10  / 2 "},
		{"3*4"},
		{"2+3*4"},
		{"(2+3)*4"},
	}

	for i, test := range tests {
		tree, _ := NewTree(test.expr)
		if tree.Flag != 1 {
			t.Errorf("неправильный флаг при expr=%s (%v)", test.expr, tree.Flag)
		}

		if !isEqual(tree.Root, testTree(i)) {
			t.Errorf("не совпадают деревья (тест %v)", i)
		}
	}
}

func testTree(id int) *Node {
	if id == 0 {
		root := &Node{Val: uint8('+'), Flag: 2}
		left := &Node{Val: float64(2), Parent: root, Flag: 3}
		right := &Node{Val: float64(2), Parent: root, Flag: 3}
		root.Left = left
		root.Right = right
		return root
	}
	if id == 1 {
		root := &Node{Val: uint8('-'), Flag: 2}
		left := &Node{Val: float64(5), Parent: root, Flag: 3}
		right := &Node{Val: float64(3), Parent: root, Flag: 3}
		root.Left = left
		root.Right = right
		return root
	}
	if id == 2 {
		root := &Node{Val: uint8('/'), Flag: 2}
		left := &Node{Val: float64(10), Parent: root, Flag: 3}
		right := &Node{Val: float64(2), Parent: root, Flag: 3}
		root.Left = left
		root.Right = right
		return root
	}
	if id == 3 {
		root := &Node{Val: uint8('*'), Flag: 2}
		left := &Node{Val: float64(3), Parent: root, Flag: 3}
		right := &Node{Val: float64(4), Parent: root, Flag: 3}
		root.Left = left
		root.Right = right
		return root
	}
	if id == 4 {
		root := &Node{Val: uint8('+'), Flag: 1}
		left := &Node{Val: float64(2), Parent: root, Flag: 3}
		right := &Node{Val: uint8('*'), Parent: root, Flag: 2}
		root.Left = left
		root.Right = right

		left = &Node{Val: float64(3), Parent: root.Right, Flag: 3}
		right = &Node{Val: float64(4), Parent: root.Right, Flag: 3}
		root.Right.Left = left
		root.Right.Right = right

		return root
	}
	if id == 5 {
		root := &Node{Val: uint8('*'), Flag: 1}
		left := &Node{Val: uint8('+'), Parent: root, Flag: 2}
		right := &Node{Val: float64(4), Parent: root, Flag: 3}
		root.Left = left
		root.Right = right

		left = &Node{Val: float64(2), Parent: root.Left, Flag: 3}
		right = &Node{Val: float64(3), Parent: root.Left, Flag: 3}
		root.Left.Left = left
		root.Left.Right = right

		return root
	}
	return nil
}

func isEqual(exp, got *Node) bool {
	if exp == nil && got == nil {
		return true
	}
	if exp == nil || got == nil {
		return false
	}
	if exp.Val != got.Val || exp.Flag != got.Flag {
		return false
	}
	if exp.Parent == nil && got.Parent != nil || exp.Parent != nil && got.Parent == nil {
		return false
	}
	if exp.Parent != nil && exp.Parent.Val != got.Parent.Val {
		return false
	}
	return isEqual(exp.Left, got.Left) && isEqual(exp.Right, got.Right)
}
