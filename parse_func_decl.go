package main

import (
	"go/ast"
	"log"
	"strings"
)

var globalObjectMap map[string]string

func ParseFuncRev(list *ast.FieldList) (rec, obj string) {
	if len(list.List) > 1 {
		log.Fatal("rev field count greater than 1")
	}
	if len(list.List[0].Names) > 1 {
		log.Fatal("rev field's name count greater than 1")
	}
	obj = ParseExpr(list.List[0].Names[0])
	if starExpr, ok := list.List[0].Type.(*ast.StarExpr); ok {
		rec = ParseExpr(starExpr.X)
	} else {
		log.Fatal("rev's type is not StartExpr")
	}
	return rec, obj
}

func ParseFuncSignature(decl *ast.FuncDecl, objectTypeMap *ObjectTypeMap) (funcRet, funcName, funcParams, funcVars string) {
	name := decl.Name.Name
	funcType := decl.Type

	params := ParseFieldList(funcType.Params)
	var results []string
	if funcType.Results != nil {
		results = ParseFieldList(funcType.Results)
	}
	//vars := results

	funcParams = "(" + strings.Join(params, ",") + ")"
	funcName = name
	if len(results) == 0 {
		funcRet = "void"
	} else if len(results) == 1 {
		res := strings.TrimSpace(results[0])
		if strings.Contains(res, " ") {
			reses := strings.Split(res, " ")
			if len(reses) != 2 {
				log.Fatal("not key-value pair: " + strings.Join(reses, " "))
			}
			funcRet = reses[0]
			funcVars = results[0] + ";"
		} else {
			funcRet = results[0]
		}
		funcRet = results[0]
	} else if len(results) == 2 {
		includeFileMap["std::pair"] = "utility"
		//funcRet = "std::pair<" + results[0] + "," + results[1] + ">"
		res1 := strings.TrimSpace(results[0])
		res2 := strings.TrimSpace(results[1])
		if strings.Contains(res1, " ") && strings.Contains(res2, " ") {
			res1s := strings.Split(res1, " ")
			res2s := strings.Split(res2, " ")
			if len(res1s) != 2 || len(res2s) != 2 {
				log.Fatal("not key value pair: " + strings.Join(res1s, " ") + ", " + strings.Join(res2s, " "))
			}
			funcRet = "std::pair<" + res1s[0] + "," + res2s[0] + ">"
			funcVars = results[0] + ";"
			funcVars += results[1] + ";"
		} else if !strings.Contains(res1, " ") && !strings.Contains(res2, " ") {
			funcRet = "std::pair<" + results[0] + "," + results[1] + ">"
		} else {
			log.Fatal("not all key-value pair: " + results[0] + ", " + results[1])
		}
	} else {
		includeFileMap["std::tuple"] = "tuple"
		if strings.Contains(strings.TrimSpace(results[0]), " ") {
			// key-value pair
			tuple := "std::tuple<"
			for id, r := range results {
				res := strings.TrimSpace(r)
				if !strings.Contains(res, " ") {
					log.Fatal("not key-value pair: " + r)
				}
				reses := strings.Split(res, " ")
				if len(reses) != 2 {
					log.Fatal("not key-value pair: " + r)
				}

				if id == 0 {
					tuple += reses[0]
				} else {
					tuple += ", " + reses[0]
				}
				funcVars += reses[1] + ";"
			}
			tuple += ">"
			funcRet = tuple
		} else {
			// only type
			tuple := "std::tuple<"
			for id, r := range results {
				if strings.Contains(strings.TrimSpace(r), " ") {
					log.Fatal("is key-value pair: " + r)
				}
				if id == 0 {
					tuple += r
				} else {
					tuple += ", " + r
				}
			}
			tuple += ">"
			funcRet = tuple
		}
	}
	return funcRet, funcName, funcParams, funcVars
}

func ParseCommonFuncDecl(decl *ast.FuncDecl, objectTypeMap *ObjectTypeMap) []string {
	var ret []string
	funcRet, funcName, funcParams, funcVars := ParseFuncSignature(decl, objectTypeMap)
	ret = append(ret, funcRet + " " + funcName + funcParams)
	body := ParseBlockStmt(decl.Body, objectTypeMap)
	ret = append(ret, "{")
	ret = append(ret, funcVars)
	ret = append(ret, body...)
	ret = append(ret, "}")
	return ret
}

var structFuncDeclMap map[string][]string = make(map[string][]string)

func GetStructFuncDeclMap() map[string][]string {
	return structFuncDeclMap
}

var structFuncDefinitionMap map[string][]string = make(map[string][]string)

func GetStructFuncDefinitionMap() map[string][]string {
	return structFuncDefinitionMap
}

func ParseMemberFuncDecl(decl *ast.FuncDecl, objectTypeMap *ObjectTypeMap) []string {
	var ret []string
	var rev string
	var revObj string
	rev, revObj = ParseFuncRev(decl.Recv)
	funcRet, funcName, funcParams, funcVars := ParseFuncSignature(decl, objectTypeMap)
	// add struct type to function map
	structFuncDeclMap[rev] = append(structFuncDeclMap[rev], funcRet + " " + funcName + funcParams + ";")
	body := ParseBlockStmt(decl.Body, objectTypeMap)

	strBody := strings.Join(body, "\n")
	strBody = "{ " + funcVars + " " + strBody  + " }"
	strBody = strings.ReplaceAll(strBody, revObj + ".", "this->")

	structFuncDefinitionMap[rev] = append(structFuncDefinitionMap[rev],
		funcRet + " " + rev + "::" + funcName + funcParams + strBody)

	return ret
}

func ParseFuncDecl(decl *ast.FuncDecl, objectTypeMap *ObjectTypeMap) []string {
	if decl.Recv != nil {
		// member function
		return ParseMemberFuncDecl(decl, objectTypeMap)
	}
	// common function
	return ParseCommonFuncDecl(decl, objectTypeMap)
}