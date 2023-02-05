package gorm_driver_redis

import (
	"fmt"
	"github.com/pingcap/tidb/parser"
	"github.com/pingcap/tidb/parser/ast"
	"github.com/pingcap/tidb/parser/mysql"
	"github.com/pingcap/tidb/parser/test_driver"
	_ "github.com/pingcap/tidb/parser/test_driver"
	"gorm.io/gorm/clause"
	"testing"
)

func TestParseDirectWhereSQL(t *testing.T) {
	var sqlParser = parser.New()
	var mockSQL = "SELECT * FROM test_models WHERE name LIKE ?"
	stmtNode, err := sqlParser.ParseOneStmt(mockSQL, mysql.DefaultCharset, "")
	if err != nil {
		panic(fmt.Errorf("failed to parse expr as sql with err:%w", err))
	}
	var whereExpr = stmtNode.(*ast.SelectStmt).Where
	fmt.Println("where expr node:", whereExpr)
	switch whereExpr.(type) {
	case *ast.PatternLikeExpr:
		var patternLikeExpr = whereExpr.(*ast.PatternLikeExpr)
		fmt.Println(patternLikeExpr)
		//patternLikeExpr.Accept(&PatternLikeVisitor{})
		var columnNameExpr = patternLikeExpr.Expr.(*ast.ColumnNameExpr)
		//var rvStr=mockSQL[patternLikeExpr.Pattern.OriginTextPosition():]

		var rvExpr = patternLikeExpr.Pattern.(*test_driver.ValueExpr)
		var rv = rvExpr.Datum.GetValue()
		patternLikeExpr.Pattern.GetType()
		var likeClause = clause.Like{
			Column: clause.Column{
				Table: clause.CurrentTable,
				Name:  columnNameExpr.Name.String(),
			},
			Value: rv,
		}
		fmt.Println(likeClause)
	case *ast.BinaryOperationExpr:
		var binaryOpExpr = whereExpr.(*ast.BinaryOperationExpr)
		fmt.Println(binaryOpExpr)
	}
	//fast path: one like expr
	if _, ok := whereExpr.(*ast.BinaryOperationExpr); ok { //only one clause and op is like or sth

	}
}

type PatternLikeVisitor struct {
}

func (plv *PatternLikeVisitor) Enter(n ast.Node) (node ast.Node, skipChildren bool) {
	return n, false
}

func (plv *PatternLikeVisitor) Leave(n ast.Node) (node ast.Node, ok bool) {
	return n, true
}
