package lexer

import (
	"simpyl/token"
	"testing"
)

func TestNextToken(t *testing.T) {
	input := `
let five = 5;
let ten = 10;
def add(x, y) {
	x + y;
};
let result = add(five, ten);
!-/*5;
5 < 10 > 5;

if (5 < 10) {
	return true;
} else {
	return false;
}

10 == 10;
10 != 9;
"foobar"
"foo bar"
[1, 2];
{"foo": "bar"}
list.append
for i in range:
while true:
3.14
.50
val in obj
# Comment
`

	tests := []struct {
		expectedType    token.TokenType
		expectedLiteral string
	}{
		{token.NEWLINE, "\n"},
		{token.LET, "let"},
		{token.IDENT, "five"},
		{token.ASSIGN, "="},
		{token.INT, "5"},
		{token.SEMICOLON, ";"},
		{token.NEWLINE, "\n"},
		{token.LET, "let"},
		{token.IDENT, "ten"},
		{token.ASSIGN, "="},
		{token.INT, "10"},
		{token.SEMICOLON, ";"},
		{token.NEWLINE, "\n"},
		{token.FUNCTION, "def"},
		{token.IDENT, "add"},
		{token.LPAREN, "("},
		{token.IDENT, "x"},
		{token.COMMA, ","},
		{token.IDENT, "y"},
		{token.RPAREN, ")"},
		{token.LBRACE, "{"},
		{token.NEWLINE, "\n"},
		{token.TAB, "\t"},
		{token.IDENT, "x"},
		{token.PLUS, "+"},
		{token.IDENT, "y"},
		{token.SEMICOLON, ";"},
		{token.NEWLINE, "\n"},
		{token.RBRACE, "}"},
		{token.SEMICOLON, ";"},
		{token.NEWLINE, "\n"},
		{token.LET, "let"},
		{token.IDENT, "result"},
		{token.ASSIGN, "="},
		{token.IDENT, "add"},
		{token.LPAREN, "("},
		{token.IDENT, "five"},
		{token.COMMA, ","},
		{token.IDENT, "ten"},
		{token.RPAREN, ")"},
		{token.SEMICOLON, ";"},
		{token.NEWLINE, "\n"},
		{token.BANG, "!"},
		{token.MINUS, "-"},
		{token.SLASH, "/"},
		{token.ASTERISK, "*"},
		{token.INT, "5"},
		{token.SEMICOLON, ";"},
		{token.NEWLINE, "\n"},
		{token.INT, "5"},
		{token.LT, "<"},
		{token.INT, "10"},
		{token.GT, ">"},
		{token.INT, "5"},
		{token.SEMICOLON, ";"},
		{token.NEWLINE, "\n"},
		{token.NEWLINE, "\n"},
		{token.IF, "if"},
		{token.LPAREN, "("},
		{token.INT, "5"},
		{token.LT, "<"},
		{token.INT, "10"},
		{token.RPAREN, ")"},
		{token.LBRACE, "{"},
		{token.NEWLINE, "\n"},
		{token.TAB, "\t"},
		{token.RETURN, "return"},
		{token.TRUE, "true"},
		{token.SEMICOLON, ";"},
		{token.NEWLINE, "\n"},
		{token.RBRACE, "}"},
		{token.ELSE, "else"},
		{token.LBRACE, "{"},
		{token.NEWLINE, "\n"},
		{token.TAB, "\t"},
		{token.RETURN, "return"},
		{token.FALSE, "false"},
		{token.SEMICOLON, ";"},
		{token.NEWLINE, "\n"},
		{token.RBRACE, "}"},
		{token.NEWLINE, "\n"},
		{token.NEWLINE, "\n"},
		{token.INT, "10"},
		{token.EQ, "=="},
		{token.INT, "10"},
		{token.SEMICOLON, ";"},
		{token.NEWLINE, "\n"},
		{token.INT, "10"},
		{token.NOT_EQ, "!="},
		{token.INT, "9"},
		{token.SEMICOLON, ";"},
		{token.NEWLINE, "\n"},
		{token.STRING, "foobar"},
		{token.NEWLINE, "\n"},
		{token.STRING, "foo bar"},
		{token.NEWLINE, "\n"},
		{token.LBRACKET, "["},
		{token.INT, "1"},
		{token.COMMA, ","},
		{token.INT, "2"},
		{token.RBRACKET, "]"},
		{token.SEMICOLON, ";"},
		{token.NEWLINE, "\n"},
		{token.LBRACE, "{"},
		{token.STRING, "foo"},
		{token.COLON, ":"},
		{token.STRING, "bar"},
		{token.RBRACE, "}"},
		{token.NEWLINE, "\n"},
		{token.IDENT, "list"},
		{token.DOT, "."},
		{token.IDENT, "append"},
		{token.NEWLINE, "\n"},
		{token.FOR, "for"},
		{token.IDENT, "i"},
		{token.IN, "in"},
		{token.IDENT, "range"},
		{token.COLON, ":"},
		{token.NEWLINE, "\n"},
		{token.WHILE, "while"},
		{token.TRUE, "true"},
		{token.COLON, ":"},
		{token.NEWLINE, "\n"},
		{token.FLOAT, "3.14"},
		{token.NEWLINE, "\n"},
		{token.FLOAT, ".50"},
		{token.NEWLINE, "\n"},
		{token.IDENT, "val"},
		{token.IN, "in"},
		{token.IDENT, "obj"},
		{token.NEWLINE, "\n"},
		{token.EOF, ""},
	}

	l := New(input)
	for i, tt := range tests {
		tok := l.NextToken()

		if tok.Type != tt.expectedType {
			t.Fatalf("tests[%d] - tokentype wrong. expected=%q, got=%q",
				i, tt.expectedType, tok.Type)
		}
		if tok.Literal != tt.expectedLiteral {
			t.Fatalf("tests[%d] - literal wrong. expected=%q, got=%q",
				i, tt.expectedLiteral, tok.Literal)
		}
	}
}
