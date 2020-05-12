package lang

import "fmt"
import "os"
import "strings"

// #cgo CFLAGS: -std=c99
// #include "bind.h"
import "C"

//this is just a function to actionize c processes in omm
func cproc(i *int, lex []Lex, PARAM_COUNT uint, name, dir, filename string, id int) Action {

  var curLex = lex[(*i)]
  var paramExp []Lex

  pCnt := 0

  //get what is in the parenthesis
  for (*i)++; (*i) < len(lex); (*i)++ {

    if lex[(*i)].Name == "(" {
      pCnt++
    }
    if lex[(*i)].Name == ")" {
      pCnt--
    }

    paramExp = append(paramExp, lex[(*i)])

    if pCnt == 0 {
      break;
    }
  }

  //remove the parenthesis
  paramExp = paramExp[1:len(paramExp) - 1]

  cbCnt := 0
  glCnt := 0
  bCnt := 0
  pCnt = 0

  var splitParams = [][]Lex{ []Lex{} }

  for _, v := range paramExp {

    if v.Name == "{" {
      cbCnt++
    }
    if v.Name == "}" {
      cbCnt--
    }

    if v.Name == "[:" {
      glCnt++
    }
    if v.Name == ":]" {
      glCnt--
    }

    if v.Name == "[" {
      bCnt++
    }
    if v.Name == "]" {
      bCnt--
    }

    if v.Name == "(" {
      pCnt++
    }
    if v.Name == ")" {
      pCnt--
    }

    if cbCnt == 0 && glCnt == 0 && bCnt == 0 && pCnt == 0 && v.Name == "," {
      splitParams = append(splitParams, []Lex{})
      continue
    }

    splitParams[len(splitParams) - 1] = append(splitParams[len(splitParams) - 1], v)

  }

  var actionSplit [][]Action

  for _, v := range splitParams {

    if len(v) == 0 {
      continue
    }

    actionSplit = append(actionSplit, Actionizer(v, true, dir, filename))
  }

  if uint(len(actionSplit)) != PARAM_COUNT {

    //throw an error
    C.colorprint(C.CString("Error while actionizing in " + curLex.Dir + "!"), C.int(12))
    fmt.Println(" Expected", PARAM_COUNT, "argument(s), but got", len(splitParams), "instead to call", /* say the process */ name, "\n\nError occured on line", curLex.Line, "\nFound near:", strings.TrimSpace(curLex.Exp))

    //exit the process
    os.Exit(1)
  }

  len_lex := len(lex)

  var actPut Action = Action{ name, "", []string{}, []Action{}, []string{}, actionSplit, []Condition{}, id, []Action{}, []Action{}, []Action{}, [][]Action{}, [][]Action{}, make(map[string][]Action), false }

  if *i + 1 < len_lex && lex[*i + 1].Name == "." {

    cbCnt := 0
    glCnt := 0
    bCnt := 0
    pCnt := 0

    indexes := [][]Lex{ []Lex{} }

    cbCnt = 0
    glCnt = 0
    bCnt = 0
    pCnt = 0

    for o := *i + 2; o < len_lex; o++ {
      if lex[o].Name == "{" {
        cbCnt++
      }
      if lex[o].Name == "[:" {
        glCnt++
      }
      if lex[o].Name == "[" {
        bCnt++
      }
      if lex[o].Name == "(" {
        pCnt++
      }

      if lex[o].Name == "}" {
        cbCnt--
      }
      if lex[o].Name == ":]" {
        glCnt--
      }
      if lex[o].Name == "]" {
        bCnt--
      }
      if lex[o].Name == ")" {
        pCnt--
      }

      if lex[o].Name == "." {
        indexes = append(indexes, []Lex{})
      } else {

        (*i)++

        indexes[len(indexes) - 1] = append(indexes[len(indexes) - 1], lex[o])

        if cbCnt == 0 && glCnt == 0 && bCnt == 0 && pCnt == 0 {

          if o < len_lex - 1 && lex[o + 1].Name == "." {
            continue
          } else {
            break
          }

        }
      }
    }

    var putIndexes [][]Action

    for _, v := range indexes {

      v = v[1:len(v) - 1]
      putIndexes = append(putIndexes, Actionizer(v, true, dir, name))
    }

    (*i)+=3

    actPut = Action{ "expressionIndex", name, []string{}, []Action{ actPut }, []string{}, [][]Action{}, []Condition{}, 8, []Action{}, []Action{}, []Action{}, [][]Action{}, putIndexes, make(map[string][]Action), false }
  }

  return actPut
}
