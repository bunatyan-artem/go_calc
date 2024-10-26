package main

import (
	"fmt"
)

type StackFloat64 struct {
	stack []float64
}

func (S *StackFloat64) push(num float64) {
	S.stack = append(S.stack, num)
}

func (S *StackFloat64) pop() (float64, error) {
	length := S.size()
	if length > 0 {
		top, _ := S.top()
		S.stack = S.stack[:length-1]
		return top, nil
	}
	return 0, fmt.Errorf("error: Stack is empty")
}

func (S *StackFloat64) top() (float64, error) {
	length := S.size()
	if length > 0 {
		return S.stack[length-1], nil
	}
	return 0, fmt.Errorf("error: Stack is empty")
}

func (S *StackFloat64) size() int {
	return len(S.stack)
}

func (S *StackFloat64) empty() bool {
	return S.size() == 0
}

//

type StackChar struct {
	stack []uint8
}

func (S *StackChar) push(c uint8) {
	S.stack = append(S.stack, c)
}

func (S *StackChar) pop() (uint8, error) {
	length := S.size()
	if length > 0 {
		top, _ := S.top()
		S.stack = S.stack[:length-1]
		return top, nil
	}
	return 0, fmt.Errorf("error: Stack is empty")
}

func (S *StackChar) top() (uint8, error) {
	length := S.size()
	if length > 0 {
		return S.stack[length-1], nil
	}
	return 0, fmt.Errorf("error: Stack is empty")
}

func (S *StackChar) size() int {
	return len(S.stack)
}

func (S *StackChar) empty() bool {
	return S.size() == 0
}

//

func IsDigit(c uint8) bool {
	return c >= '0' && c <= '9'
}

func IsOperation(c uint8) bool {
	return c == '+' || c == '-' || c == '*' || c == '/'
}

func Priority(c uint8) uint8 {
	switch c {
	case '+':
		return 1
	case '-':
		return 1
	case '*':
		return 2
	case '/':
		return 2
	default:
		return 0
	}
}

func ElementaryCalculation(nums *StackFloat64, op uint8) error {
	num2, err1 := nums.pop()
	num1, err2 := nums.pop()

	if err1 != nil || err2 != nil {
		return fmt.Errorf("error: ElementaryCalculation was called with less than 2 nums in stack")
	}

	switch op {
	case '+':
		nums.push(num1 + num2)
		return nil
	case '-':
		nums.push(num1 - num2)
		return nil
	case '*':
		nums.push(num1 * num2)
		return nil
	case '/':
		if num2 == 0 {
			return fmt.Errorf("error: division by zero")
		}
		nums.push(num1 / num2)
		return nil
	case '(':
		return fmt.Errorf("error: '(' without ')'")
	}
	return fmt.Errorf("error: unknown operation")
}

func Calc(expression string) (float64, error) {
	if expression == "" {
		return 0, fmt.Errorf("error: empty expression")
	}
	var nums StackFloat64
	var ops StackChar

	var lastC uint8 = ' '
	for i := 0; i < len(expression); i++ {
		c := expression[i]
		if c == ' ' {
			continue
		} else if c == '(' {
			if IsDigit(lastC) {
				return 0, fmt.Errorf("error: num before '('")
			}
			ops.push('(')
		} else if c == ')' {
			if !IsDigit(lastC) && lastC != ')' {
				return 0, fmt.Errorf("error: wrong input (%d)", i)
			}
			for {
				op, err := ops.pop()
				if err != nil {
					return 0, fmt.Errorf("error: ')' without '('")
				}

				if op == '(' {
					break
				}

				err = ElementaryCalculation(&nums, op)
				if err != nil {
					return 0, err
				}
			}
		} else if IsOperation(c) {
			if !IsDigit(lastC) && lastC != ')' {
				return 0, fmt.Errorf("error: wrong input (%d)", i)
			}
			for op, _ := ops.top(); !ops.empty() && Priority(op) >= Priority(c); {
				op, _ := ops.pop()
				err := ElementaryCalculation(&nums, op)
				if err != nil {
					return 0, err
				}
			}
			ops.push(c)
		} else if IsDigit(c) {
			num := 0.0
			for i < len(expression) && IsDigit(expression[i]) {
				num *= 10
				num += float64(expression[i] - '0')
				i++
			}
			i--
			nums.push(num)
		} else {
			return 0, fmt.Errorf("error: invalid symbol (%d)", i)
		}
		lastC = c
	}

	if IsOperation(lastC) {
		return 0, fmt.Errorf("error: wrong input at the end")
	}

	for !ops.empty() {
		op, _ := ops.pop()
		err := ElementaryCalculation(&nums, op)
		if err != nil {
			return 0, err
		}
	}
	return nums.top()
}
