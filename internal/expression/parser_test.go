package expression

import (
	"testing"
)

func TestParse(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantExpr int
		wantErr  bool
	}{
		{
			name:     "simple expression",
			input:    "${{ event.file.path }}",
			wantExpr: 1,
		},
		{
			name:     "no expression",
			input:    "just a string",
			wantExpr: 0,
		},
		{
			name:     "multiple expressions",
			input:    "${{ event.cwd }} and ${{ event.timestamp }}",
			wantExpr: 2,
		},
		{
			name:     "expression with function",
			input:    "${{ contains(event.file.path, 'test') }}",
			wantExpr: 1,
		},
		{
			name:     "expression with whitespace",
			input:    "${{   event.file.path   }}",
			wantExpr: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			exprs, err := Parse(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("Parse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if len(exprs) != tt.wantExpr {
				t.Errorf("Parse() got %d expressions, want %d", len(exprs), tt.wantExpr)
			}
		})
	}
}

func TestContainsExpression(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{"${{ event.file }}", true},
		{"no expression here", false},
		{"partial ${{ incomplete", false},
		{"${{ a }} ${{ b }}", true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			if got := ContainsExpression(tt.input); got != tt.want {
				t.Errorf("ContainsExpression() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestExtractExpressions(t *testing.T) {
	input := "Hello ${{ name }}, your path is ${{ event.file.path }}"
	exprs := ExtractExpressions(input)

	if len(exprs) != 2 {
		t.Errorf("Expected 2 expressions, got %d", len(exprs))
	}

	if exprs[0] != "name" {
		t.Errorf("Expected first expression to be 'name', got '%s'", exprs[0])
	}

	if exprs[1] != "event.file.path" {
		t.Errorf("Expected second expression to be 'event.file.path', got '%s'", exprs[1])
	}
}

func TestTokenize(t *testing.T) {
	tests := []struct {
		name    string
		expr    string
		wantLen int
		wantErr bool
	}{
		{
			name:    "simple identifier",
			expr:    "event",
			wantLen: 2, // identifier + EOF
		},
		{
			name:    "property access",
			expr:    "event.file.path",
			wantLen: 6, // event . file . path EOF
		},
		{
			name:    "function call",
			expr:    "contains(a, b)",
			wantLen: 7, // contains ( a , b ) EOF
		},
		{
			name:    "string literal",
			expr:    "'hello world'",
			wantLen: 2, // string + EOF
		},
		{
			name:    "number",
			expr:    "42",
			wantLen: 2, // number + EOF
		},
		{
			name:    "comparison",
			expr:    "a == b",
			wantLen: 4, // a == b EOF
		},
		{
			name:    "boolean operators",
			expr:    "a && b || c",
			wantLen: 6, // a && b || c EOF
		},
		{
			name:    "escaped quote",
			expr:    "'it''s'",
			wantLen: 2, // string + EOF
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tokens, err := tokenize(tt.expr)
			if (err != nil) != tt.wantErr {
				t.Errorf("tokenize() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if len(tokens) != tt.wantLen {
				t.Errorf("tokenize() got %d tokens, want %d", len(tokens), tt.wantLen)
				for i, tok := range tokens {
					t.Logf("Token %d: type=%v value=%q", i, tok.Type, tok.Value)
				}
			}
		})
	}
}

func TestReadString(t *testing.T) {
	tests := []struct {
		input   string
		want    string
		wantLen int
		wantErr bool
	}{
		{"'hello'", "hello", 7, false},
		{"'it''s ok'", "it's ok", 10, false},
		{"'empty'more", "empty", 7, false},
		{"'unterminated", "", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, gotLen, err := readString([]rune(tt.input))
			if (err != nil) != tt.wantErr {
				t.Errorf("readString() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("readString() = %q, want %q", got, tt.want)
			}
			if gotLen != tt.wantLen {
				t.Errorf("readString() length = %d, want %d", gotLen, tt.wantLen)
			}
		})
	}
}

func TestReadNumber(t *testing.T) {
	tests := []struct {
		input   string
		want    string
		wantLen int
	}{
		{"42", "42", 2},
		{"3.14", "3.14", 4},
		{"-7", "-7", 2},
		{"1e10", "1e10", 4},
		{"2.5e-3", "2.5e-3", 6},
		{"100abc", "100", 3},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, gotLen := readNumber([]rune(tt.input))
			if got != tt.want {
				t.Errorf("readNumber() = %q, want %q", got, tt.want)
			}
			if gotLen != tt.wantLen {
				t.Errorf("readNumber() length = %d, want %d", gotLen, tt.wantLen)
			}
		})
	}
}
