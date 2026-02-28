package expression

import (
	"fmt"
	"regexp"
	"strings"
)

// ExpressionPattern matches ${{ ... }} expressions
var ExpressionPattern = regexp.MustCompile(`\$\{\{\s*(.*?)\s*\}\}`)

// Token represents a parsed token in an expression
type Token struct {
	Type  TokenType
	Value string
}

// TokenType identifies the type of token
type TokenType int

const (
	TokenLiteral TokenType = iota
	TokenIdentifier
	TokenNumber
	TokenString
	TokenOperator
	TokenLeftParen
	TokenRightParen
	TokenLeftBracket
	TokenRightBracket
	TokenDot
	TokenComma
	TokenEOF
)

// Expression represents a parsed expression
type Expression struct {
	Raw    string   // Original expression string
	Tokens []Token  // Parsed tokens
}

// Parse extracts and parses expressions from a string
// Returns the original string with expression placeholders and a map of expressions
func Parse(input string) ([]Expression, error) {
	matches := ExpressionPattern.FindAllStringSubmatch(input, -1)
	if len(matches) == 0 {
		return nil, nil
	}

	var expressions []Expression
	for _, match := range matches {
		if len(match) < 2 {
			continue
		}
		expr := Expression{
			Raw: match[1],
		}
		tokens, err := tokenize(match[1])
		if err != nil {
			return nil, fmt.Errorf("failed to tokenize expression '%s': %w", match[1], err)
		}
		expr.Tokens = tokens
		expressions = append(expressions, expr)
	}

	return expressions, nil
}

// ContainsExpression checks if a string contains any expressions
func ContainsExpression(input string) bool {
	return ExpressionPattern.MatchString(input)
}

// ExtractExpressions returns all expression strings from input
func ExtractExpressions(input string) []string {
	matches := ExpressionPattern.FindAllStringSubmatch(input, -1)
	var exprs []string
	for _, match := range matches {
		if len(match) >= 2 {
			exprs = append(exprs, match[1])
		}
	}
	return exprs
}

// ReplaceExpressions replaces all expressions in a string using a replacement function
func ReplaceExpressions(input string, replacer func(expr string) (string, error)) (string, error) {
	result := input
	matches := ExpressionPattern.FindAllStringSubmatchIndex(input, -1)
	
	// Process in reverse order to maintain correct indices
	for i := len(matches) - 1; i >= 0; i-- {
		match := matches[i]
		if len(match) < 4 {
			continue
		}
		
		fullStart, fullEnd := match[0], match[1]
		exprStart, exprEnd := match[2], match[3]
		
		expr := strings.TrimSpace(input[exprStart:exprEnd])
		replacement, err := replacer(expr)
		if err != nil {
			return "", fmt.Errorf("failed to evaluate expression '%s': %w", expr, err)
		}
		
		result = result[:fullStart] + replacement + result[fullEnd:]
	}
	
	return result, nil
}

// tokenize breaks an expression string into tokens
func tokenize(expr string) ([]Token, error) {
	var tokens []Token
	runes := []rune(expr)
	i := 0

	for i < len(runes) {
		ch := runes[i]

		// Skip whitespace
		if ch == ' ' || ch == '\t' || ch == '\n' || ch == '\r' {
			i++
			continue
		}

		// Single-character tokens
		switch ch {
		case '(':
			tokens = append(tokens, Token{Type: TokenLeftParen, Value: "("})
			i++
			continue
		case ')':
			tokens = append(tokens, Token{Type: TokenRightParen, Value: ")"})
			i++
			continue
		case '[':
			tokens = append(tokens, Token{Type: TokenLeftBracket, Value: "["})
			i++
			continue
		case ']':
			tokens = append(tokens, Token{Type: TokenRightBracket, Value: "]"})
			i++
			continue
		case '.':
			tokens = append(tokens, Token{Type: TokenDot, Value: "."})
			i++
			continue
		case ',':
			tokens = append(tokens, Token{Type: TokenComma, Value: ","})
			i++
			continue
		}

		// Operators
		if isOperatorStart(ch) {
			op, length := readOperator(runes[i:])
			if op != "" {
				tokens = append(tokens, Token{Type: TokenOperator, Value: op})
				i += length
				continue
			}
		}

		// String literals
		if ch == '\'' {
			str, length, err := readString(runes[i:])
			if err != nil {
				return nil, err
			}
			tokens = append(tokens, Token{Type: TokenString, Value: str})
			i += length
			continue
		}

		// Numbers
		if isDigit(ch) || (ch == '-' && i+1 < len(runes) && isDigit(runes[i+1])) {
			num, length := readNumber(runes[i:])
			tokens = append(tokens, Token{Type: TokenNumber, Value: num})
			i += length
			continue
		}

		// Identifiers
		if isIdentifierStart(ch) {
			ident, length := readIdentifier(runes[i:])
			tokens = append(tokens, Token{Type: TokenIdentifier, Value: ident})
			i += length
			continue
		}

		return nil, fmt.Errorf("unexpected character '%c' at position %d", ch, i)
	}

	tokens = append(tokens, Token{Type: TokenEOF, Value: ""})
	return tokens, nil
}

func isOperatorStart(ch rune) bool {
	return ch == '!' || ch == '=' || ch == '<' || ch == '>' || ch == '&' || ch == '|' || ch == '+' || ch == '-' || ch == '*' || ch == '/'
}

func readOperator(runes []rune) (string, int) {
	if len(runes) >= 2 {
		two := string(runes[:2])
		switch two {
		case "==", "!=", "<=", ">=", "&&", "||":
			return two, 2
		}
	}
	if len(runes) >= 1 {
		one := string(runes[0])
		switch one {
		case "!", "<", ">", "+", "-", "*", "/":
			return one, 1
		}
	}
	return "", 0
}

func readString(runes []rune) (string, int, error) {
	if runes[0] != '\'' {
		return "", 0, fmt.Errorf("expected string to start with single quote")
	}
	
	var sb strings.Builder
	i := 1
	for i < len(runes) {
		if runes[i] == '\'' {
			// Check for escaped quote
			if i+1 < len(runes) && runes[i+1] == '\'' {
				sb.WriteRune('\'')
				i += 2
				continue
			}
			// End of string
			return sb.String(), i + 1, nil
		}
		sb.WriteRune(runes[i])
		i++
	}
	return "", 0, fmt.Errorf("unterminated string")
}

func readNumber(runes []rune) (string, int) {
	var sb strings.Builder
	i := 0
	
	// Handle negative
	if runes[i] == '-' {
		sb.WriteRune('-')
		i++
	}
	
	// Integer part
	for i < len(runes) && isDigit(runes[i]) {
		sb.WriteRune(runes[i])
		i++
	}
	
	// Decimal part
	if i < len(runes) && runes[i] == '.' {
		sb.WriteRune('.')
		i++
		for i < len(runes) && isDigit(runes[i]) {
			sb.WriteRune(runes[i])
			i++
		}
	}
	
	// Exponent
	if i < len(runes) && (runes[i] == 'e' || runes[i] == 'E') {
		sb.WriteRune(runes[i])
		i++
		if i < len(runes) && (runes[i] == '+' || runes[i] == '-') {
			sb.WriteRune(runes[i])
			i++
		}
		for i < len(runes) && isDigit(runes[i]) {
			sb.WriteRune(runes[i])
			i++
		}
	}
	
	return sb.String(), i
}

func readIdentifier(runes []rune) (string, int) {
	var sb strings.Builder
	i := 0
	for i < len(runes) && isIdentifierChar(runes[i]) {
		sb.WriteRune(runes[i])
		i++
	}
	return sb.String(), i
}

func isDigit(ch rune) bool {
	return ch >= '0' && ch <= '9'
}

func isIdentifierStart(ch rune) bool {
	return (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') || ch == '_'
}

func isIdentifierChar(ch rune) bool {
	return isIdentifierStart(ch) || isDigit(ch) || ch == '-'
}
