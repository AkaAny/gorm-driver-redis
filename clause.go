package gorm_driver_redis

import (
	"fmt"
	"github.com/pingcap/tidb/parser"
	"github.com/pingcap/tidb/parser/ast"
	"github.com/pingcap/tidb/parser/mysql"
	"github.com/pingcap/tidb/parser/opcode"
	"github.com/pingcap/tidb/parser/test_driver"
	_ "github.com/pingcap/tidb/parser/test_driver"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func ParseWhereClauses(tx *gorm.DB, whereClause clause.Where) (clause.Where, error) {
	var sqlParser = parser.New()
	//utils.Map[clause.Expression,clause.Expression]()
	//sqlParser.ParseOneStmt()
	var replaceIndexClauseMap = make(map[int]clause.Expression)
	for i, exprInterface := range whereClause.Exprs {
		exprClause, ok := exprInterface.(clause.Expr) //str where
		if !ok {
			continue
		}
		fmt.Println("parse:", i, exprInterface)
		//it is a where clause so we can pretend that we have a full sql
		var mockSQL = fmt.Sprintf("SELECT * FROM %s WHERE %s",
			tx.Statement.Schema.Table, exprClause.SQL)
		fmt.Println("mock sql:", mockSQL)
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
			var likeClause = parsePatternLikeExpr(exprClause, mockSQL, patternLikeExpr)
			replaceIndexClauseMap[i] = likeClause
		case *ast.BinaryOperationExpr:
			var binaryOpExpr = whereExpr.(*ast.BinaryOperationExpr)
			fmt.Println(binaryOpExpr)
			//we now do not consider multiple cond clause in one where func
			switch binaryOpExpr.Op {
			case opcode.EQ:
				var columnNameExpr = binaryOpExpr.L.(*ast.ColumnNameExpr)
				var rv interface{} = getRVFromExpr(exprClause.Vars[0], binaryOpExpr.R)
				var eqClause = clause.Eq{
					Column: clause.Column{
						Table: clause.CurrentTable,
						Name:  columnNameExpr.Name.String(),
					},
					Value: rv,
				}
				replaceIndexClauseMap[i] = eqClause
			default:
				panic("unsupported binary op")
			}
		}
		//var selectNode=stmtNode
		fmt.Println(stmtNode)
	}
	for i, exprInterface := range replaceIndexClauseMap {
		whereClause.Exprs[i] = exprInterface
	}
	return whereClause, nil
}

func parsePatternLikeExpr(exprClause clause.Expr, mockSQL string, patternLikeExpr *ast.PatternLikeExpr) clause.Like {
	var columnNameExpr = patternLikeExpr.Expr.(*ast.ColumnNameExpr)
	var rv interface{} = getRVFromExpr(exprClause.Vars[0], patternLikeExpr.Pattern)
	var likeClause = clause.Like{
		Column: clause.Column{
			Table: clause.CurrentTable,
			Name:  columnNameExpr.Name.String(),
		},
		Value: rv,
	}
	return likeClause
}

func getRVFromExpr(varFromGormClause interface{}, rvExpr ast.ExprNode) (rv interface{}) {
	switch rvExpr.(type) {
	case *test_driver.ParamMarkerExpr:
		rv = varFromGormClause
	case *test_driver.ValueExpr:
		var valueExpr = rvExpr.(*test_driver.ValueExpr)
		rv = valueExpr.Datum.GetValue()
	default:
		panic("unsupported rv expr type")
	}
	return rv
}
